package urknall

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pubsub"
)

type checksumTree map[string]map[string]struct{}

type Provisioner interface {
	ProvisionRunlist(*Package, checksumTree) error
	BuildChecksumTree() (checksumTree, error)
}

// Provision the given list of runlists.
func provisionRunlists(runLists []*Package, runner *Runner) (e error) {
	ct, e := buildChecksumTree(runner)
	if e != nil {
		return e
	}

	for i := range runLists {
		rl := runLists[i]
		m := &Message{key: MessageRunlistsProvision, runlist: rl}
		m.publish("started")
		if e = provisionRunlist(runner, rl, ct); e != nil {
			m.publishError(e)
			return e
		}
		m.Publish("finished")
	}
	return nil
}

func (prov *provisioner) prepareHost() (e error) {
	con, e := prov.sshClient.Connection()
	if e != nil {
		return e
	}

	if e := executeCommand(con, fmt.Sprintf(`grep "^%s:" /etc/group | grep %s`, ukGROUP, prov.host.user())); e != nil {
		// If user is missing the group, create group (if necessary), add user and restart ssh connection.
		cmds := []string{
			fmt.Sprintf(`{ grep -e '^%[1]s:' /etc/group > /dev/null || { groupadd %[1]s; }; }`, ukGROUP),
			fmt.Sprintf(`{ [[ -d %[1]s ]] || { mkdir -p -m 2775 %[1]s && chgrp %[2]s %[1]s; }; }`, ukCACHEDIR, ukGROUP),
			fmt.Sprintf("usermod -a -G %s %s", ukGROUP, prov.host.user()),
		}

		if e := executeCommand(con, fmt.Sprintf(`sudo bash -c "%s"`, strings.Join(cmds, " && "))); e != nil {
			return fmt.Errorf("failed to initiate user %q for provisioning: %s", prov.host.user(), e)
		}

		// Restarting the connection is required to make sure the user's new group is added properly.
		prov.sshClient.Conn.Close()
		prov.sshClient.Conn = nil
	}
	return nil
}

func (prov *provisioner) ProvisionRunlist(rl *Runlist, ct checksumTree) (e error) {
	tasks := prov.buildTasksForRunlist(rl)

	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", rl.name)

	var found bool
	var checksumHash map[string]struct{}
	if checksumHash, found = ct[rl.name]; !found {
		ct[rl.name] = map[string]struct{}{}
		checksumHash = ct[rl.name]

		// Create checksum dir and set group bit (all new files will inherit the directory's group). This allows for
		// different users (being part of that group) to create, modify and delete the contained checksum and log files.
		createChecksumDirCmd := fmt.Sprintf("mkdir -m2775 -p %s", checksumDir)
		if prov.host.isSudoRequired() {
			createChecksumDirCmd = fmt.Sprintf(`sudo %s`, createChecksumDirCmd)
		}
		r, e := prov.sshClient.Execute(createChecksumDirCmd)
		if e != nil {
			return fmt.Errorf(r.Stderr() + ": " + e.Error())
		}
	}

	for i := range tasks {
		task := tasks[i]
		m := message(pubsub.MessageRunlistsProvisionTask, prov.host, rl)
		m.TaskChecksum = task.checksum
		m.Message = task.command.Logging()
		if _, found := checksumHash[task.checksum]; found { // Task is cached.
			m.ExecStatus = pubsub.StatusCached
			m.Publish("finished")
			delete(checksumHash, task.checksum) // Delete checksums of cached tasks from hash.
			continue
		}

		if len(checksumHash) > 0 { // All remaining checksums are invalid, as something changed.
			if e = prov.cleanUpRemainingCachedEntries(checksumDir, checksumHash); e != nil {
				return e
			}
			checksumHash = make(map[string]struct{})
		}
		m.ExecStatus = pubsub.StatusExecStart
		m.Publish("started")
		e = prov.runTask(task, checksumDir)
		if e != nil {
			m.PublishError(e)
			return e
		}
		m.ExecStatus = pubsub.StatusExecFinished
		m.Publish("finished")
	}

	return nil
}

func (prov *provisioner) runTask(task *taskData, checksumDir string) (e error) {
	if prov.provisionOptions.DryRun {
		return nil
	}

	con, e := prov.sshClient.Connection()
	if e != nil {
		return e
	}
	runner := &remoteTaskRunner{clientConn: con, task: task, host: prov.host, dir: checksumDir}
	return runner.run()
}

func (prov *provisioner) BuildChecksumTree() (ct checksumTree, e error) {
	ct = checksumTree{}

	rsp, e := prov.sshClient.Execute(fmt.Sprintf(`[[ -d %[1]s ]] && find %[1]s -type f -name \*.done`, ukCACHEDIR))
	if e != nil {
		return nil, e
	}
	for _, line := range strings.Split(rsp.Stdout(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		pkgname := filepath.Dir(strings.TrimPrefix(line, ukCACHEDIR+"/"))
		checksum := strings.TrimSuffix(filepath.Base(line), ".done")
		if len(checksum) != 64 {
			return nil, fmt.Errorf("invalid checksum %q found for package %q", checksum, pkgname)
		}
		if _, found := ct[pkgname]; !found {
			ct[pkgname] = map[string]struct{}{}
		}
		ct[pkgname][checksum] = struct{}{}
	}

	return ct, nil
}

func (prov *provisioner) cleanUpRemainingCachedEntries(checksumDir string, checksumHash map[string]struct{}) (e error) {
	invalidCacheEntries := make([]string, 0, len(checksumHash))
	for k, _ := range checksumHash {
		invalidCacheEntries = append(invalidCacheEntries, fmt.Sprintf("%s.done", k))
	}
	if prov.provisionOptions.DryRun {
		m := message(pubsub.MessageCleanupCacheEntries, prov.host, nil)
		m.InvalidatedCacheEntries = invalidCacheEntries
		m.Publish("dryrun")
	} else {
		cmd := fmt.Sprintf("cd %s && rm -f *.failed %s", checksumDir, strings.Join(invalidCacheEntries, " "))
		m := message(pubsub.MessageUrknallInternal, prov.host, nil)
		m.Message = cmd
		m.Publish("started")
		result, _ := prov.sshClient.Execute(cmd)
		m.SshResult = result
		m.Publish("finished")
	}
	return nil
}

type taskData struct {
	command  cmd.Command // The command to be executed.
	checksum string      // The checksum of the command.
	runlist  *Runlist
}

func (data *taskData) Command() cmd.Command {
	return data.command
}

func (prov *provisioner) buildTasksForRunlist(rl *Runlist) (tasks []*taskData) {
	tasks = make([]*taskData, 0, len(rl.commands))

	cmdHash := sha256.New()
	for i := range rl.commands {
		rawCmd := rl.commands[i].Shell()
		cmdHash.Write([]byte(rawCmd))

		task := &taskData{runlist: rl, command: rl.commands[i], checksum: fmt.Sprintf("%x", cmdHash.Sum(nil))}
		tasks = append(tasks, task)
	}
	return tasks
}
