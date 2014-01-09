package urknall

import (
	"crypto/sha256"
	"fmt"
	"github.com/dynport/gossh"
	"github.com/dynport/urknall/cmd"
	"path"
	"strings"
	"time"
)

type ProvisionOptions struct {
	DryRun bool
}

type sshClient struct {
	client           *gossh.Client
	host             *Host
	provisionOptions ProvisionOptions
}

func newSSHClient(host *Host, opts *ProvisionOptions) (client *sshClient) {
	if opts == nil {
		opts = &ProvisionOptions{}
	}
	c := gossh.New(host.IP, host.user())
	if host.Password != "" {
		c.SetPassword(host.Password)
	}
	return &sshClient{host: host, client: c, provisionOptions: *opts}
}

func (sc *sshClient) provision() (e error) {
	if e = sc.host.precompileRunlists(); e != nil {
		return e
	}

	if e = sc.prepareHost(); e != nil {
		return e
	}

	return provisionRunlists(sc.host.runlists(), sc.provisionRunlist)
}

func (sc *sshClient) prepareHost() (e error) {
	if !sc.host.isSudoRequired() { // nothing required to do if root is used directly
		return nil
	}

	con, e := sc.client.Connection()
	if e != nil {
		return e
	}

	if e := executeCommand(con, fmt.Sprintf(`grep "^%s:" /etc/group | grep %s`, ukGROUP, sc.host.User)); e != nil {
		// If user is missing the group, create group (if necessary), add user and restart ssh connection.
		cmds := []string{
			fmt.Sprintf(`{ grep -e '^%[1]s:' /etc/group > /dev/null || { groupadd %[1]s; }; }`, ukGROUP),
			fmt.Sprintf(`{ [[ -d %[1]s ]] || { mkdir -p -m 2775 %[1]s && chgrp %[2]s %[1]s; }; }`, ukCACHEDIR, ukGROUP),
			fmt.Sprintf("usermod -a -G %s %s", ukGROUP, sc.host.User),
		}

		if e := executeCommand(con, fmt.Sprintf(`sudo bash -c "%s"`, strings.Join(cmds, " && "))); e != nil {
			return fmt.Errorf("failed to initiate user %q for provisioning: %s", sc.host.User, e)
		}

		// Restarting the connection is required to make sure the user's new group is added properly.
		sc.client.Conn.Close()
		sc.client.Conn = nil
	}
	return nil
}

func (sc *sshClient) provisionRunlist(rl *Runlist) (e error) {
	tasks := sc.buildTasksForRunlist(rl)

	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", rl.name)

	checksumHash, e := sc.buildChecksumHash(checksumDir)
	if e != nil {
		return fmt.Errorf("failed to build checksum hash: %s", e.Error())
	}

	for i := range tasks {
		task := tasks[i]
		logMsg := task.command.Logging()
		m := &Message{key: MessageRunlistsProvisionTask, task: task, message: logMsg, host: sc.host, runlist: rl}
		if _, found := checksumHash[task.checksum]; found { // Task is cached.
			m.execStatus = statusCached
			m.publish("finished")
			delete(checksumHash, task.checksum) // Delete checksums of cached tasks from hash.
			continue
		}

		if len(checksumHash) > 0 { // All remaining checksums are invalid, as something changed.
			if e = sc.cleanUpRemainingCachedEntries(checksumDir, checksumHash); e != nil {
				return e
			}
			checksumHash = make(map[string]struct{})
		}
		m.execStatus = statusExecStart
		m.publish("started")
		e = sc.runTask(task, checksumDir)
		m.error_ = e
		m.execStatus = statusExecFinished
		m.publish("finished")
		if e != nil {
			return e
		}
	}

	return nil
}

func newDebugWriter(host *Host, task *taskData) func(i ...interface{}) {
	started := time.Now()
	return func(i ...interface{}) {
		parts := strings.SplitN(fmt.Sprint(i...), "\t", 3)
		if len(parts) == 3 {
			stream, line := parts[1], parts[2]
			var runlist *Runlist = nil
			if task != nil {
				runlist = task.runlist
			}
			m := &Message{key: "task.io", host: host, stream: stream, task: task, line: line, runlist: runlist, totalRuntime: time.Now().Sub(started)}
			m.publish(stream)
		}
	}
}

func (sc *sshClient) runTask(task *taskData, checksumDir string) (e error) {
	if sc.provisionOptions.DryRun {
		return nil
	}

	sCmd := fmt.Sprintf("bash <<EOF_RUNTASK\nset -xe\n%s\nEOF_RUNTASK\n", task.command.Shell())
	if sc.host.isSudoRequired() {
		sCmd = fmt.Sprintf("sudo %s", sCmd)
	}
	con, e := sc.client.Connection()
	if e != nil {
		return e
	}
	runner := &remoteTaskRunner{clientConn: con, cmd: sCmd, task: task, host: sc.host, dir: checksumDir}
	return runner.run()
}

func (sc *sshClient) buildChecksumHash(checksumDir string) (checksumMap map[string]struct{}, e error) {
	// Create checksum dir and set group bit (all new files will inherit the directory's group). This allows for
	// different users (being part of that group) to create, modify and delete the contained checksum and log files.
	createChecksumDirCmd := fmt.Sprintf("mkdir -m2775 -p %s", checksumDir)
	if sc.host.isSudoRequired() {
		createChecksumDirCmd = fmt.Sprintf(`sudo %s`, createChecksumDirCmd)
	}
	r, e := sc.client.Execute(createChecksumDirCmd)
	if e != nil {
		return nil, fmt.Errorf(r.Stderr() + ": " + e.Error())
	}

	checksums := []string{}

	rsp, e := sc.client.Execute(fmt.Sprintf(`for f in %s/*.done; do if [[ -f $f ]]; then echo -n "$f "; fi; done`, checksumDir))
	if e != nil {
		return nil, e
	}
	for _, checksumFile := range strings.Fields(rsp.Stdout()) {
		checksum := strings.TrimSuffix(path.Base(checksumFile), ".done")
		checksums = append(checksums, checksum)
	}

	checksumMap = make(map[string]struct{})
	for i := range checksums {
		if len(checksums[i]) != 64 {
			return nil, fmt.Errorf("invalid checksum %q found in %q", checksums, checksumDir)
		}
		checksumMap[checksums[i]] = struct{}{}
	}
	return checksumMap, nil
}

func (sc *sshClient) cleanUpRemainingCachedEntries(checksumDir string, checksumHash map[string]struct{}) (e error) {
	invalidCacheEntries := make([]string, 0, len(checksumHash))
	for k, _ := range checksumHash {
		invalidCacheEntries = append(invalidCacheEntries, fmt.Sprintf("%s.done", k))
	}
	if sc.provisionOptions.DryRun {
		(&Message{key: MessageCleanupCacheEntries, invalidatedCachentries: invalidCacheEntries, host: sc.host}).publish(".dryrun")
	} else {
		cmd := fmt.Sprintf("cd %s && rm -f *.failed %s", checksumDir, strings.Join(invalidCacheEntries, " "))
		m := &Message{command: cmd, host: sc.host, key: MessageUrknallInternal}
		m.publish("started")
		result, _ := sc.client.Execute(cmd)
		m.sshResult = result
		m.publish("finished")
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

func (sc *sshClient) buildTasksForRunlist(rl *Runlist) (tasks []*taskData) {
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
