package urknall

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pubsub"
)

type checksumTree map[string]map[string]struct{}

type Provisioner interface {
	ProvisionRunlist(*Task, checksumTree) error
	BuildChecksumTree() (checksumTree, error)
}

func buildChecksumTree(runner *Runner) (ct checksumTree, e error) {
	ct = checksumTree{}

	cmd, e := runner.Command(fmt.Sprintf(`[ -d %[1]s ] && find %[1]s -type f -name \*.done`, ukCACHEDIR))
	if e != nil {
		return nil, e
	}
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}

	cmd.SetStdout(out)
	cmd.SetStderr(err)

	if e := cmd.Run(); e != nil {
		return nil, e
	}
	for _, line := range strings.Split(out.String(), "\n") {
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

// Provision the given list of runlists.
func (runner *Runner) provision(list *Package) (e error) {
	ct, e := buildChecksumTree(runner)
	if e != nil {
		return e
	}

	for i := range list.items {
		rl := list.items[i]
		m := &pubsub.Message{Key: pubsub.MessageRunlistsProvision, Hostname: runner.Hostname()}
		m.Publish("started")
		if e = provisionRunlist(runner, rl, ct); e != nil {
			m.PublishError(e)
			return e
		}
		m.Publish("finished")
	}
	return nil
}

func provisionRunlist(runner *Runner, item *PackageListItem, ct checksumTree) (e error) {
	tasks := item.Package.tasks()

	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", item.Key)

	var found bool
	var checksumHash map[string]struct{}
	if checksumHash, found = ct[item.Key]; !found {
		ct[item.Key] = map[string]struct{}{}
		checksumHash = ct[item.Key]

		// Create checksum dir and set group bit (all new files will inherit the directory's group). This allows for
		// different users (being part of that group) to create, modify and delete the contained checksum and log files.
		createChecksumDirCmd := fmt.Sprintf("mkdir -m2775 -p %s", checksumDir)

		cmd, e := runner.Command(createChecksumDirCmd)
		if e != nil {
			return e
		}
		err := &bytes.Buffer{}

		cmd.SetStderr(err)

		if e := cmd.Run(); e != nil {
			return fmt.Errorf(err.String() + ": " + e.Error())
		}
	}

	for i := range tasks {
		task := tasks[i]
		logMsg := task.command.Logging()
		m := &pubsub.Message{Key: pubsub.MessageRunlistsProvisionTask, TaskChecksum: task.checksum, Message: logMsg, Hostname: runner.Hostname(), RunlistName: item.Key}
		if _, found := checksumHash[task.checksum]; found { // Task is cached.
			m.ExecStatus = pubsub.StatusCached
			m.Publish("finished")
			delete(checksumHash, task.checksum) // Delete checksums of cached tasks from hash.
			continue
		}

		if len(checksumHash) > 0 { // All remaining checksums are invalid, as something changed.
			if e = cleanUpRemainingCachedEntries(runner, checksumDir, checksumHash); e != nil {
				return e
			}
			checksumHash = make(map[string]struct{})
		}
		m.ExecStatus = pubsub.StatusExecStart
		m.Publish("started")
		e = runTask(runner, task, checksumDir)
		m.Error = e
		m.ExecStatus = pubsub.StatusExecFinished
		m.Publish("finished")
		if e != nil {
			return e
		}
	}

	return nil
}

func runTask(runner *Runner, task *taskData, checksumDir string) (e error) {
	if runner.DryRun {
		return nil
	}
	sCmd := fmt.Sprintf("sh -x -e -c %q", task.command.Shell())
	for _, env := range runner.Env {
		sCmd = env + " " + sCmd
	}
	r := &remoteTaskRunner{Runner: runner, cmd: sCmd, task: task, dir: checksumDir}
	return r.run()
}

type taskData struct {
	command  cmd.Command // The command to be executed.
	checksum string      // The checksum of the command.
	runlist  *Task
}

func (data *taskData) Command() cmd.Command {
	return data.command
}

func cleanUpRemainingCachedEntries(runner *Runner, checksumDir string, checksumHash map[string]struct{}) (e error) {
	invalidCacheEntries := make([]string, 0, len(checksumHash))
	for k, _ := range checksumHash {
		invalidCacheEntries = append(invalidCacheEntries, fmt.Sprintf("%s.done", k))
	}
	if runner.DryRun {
		(&pubsub.Message{Key: pubsub.MessageCleanupCacheEntries, InvalidatedCacheEntries: invalidCacheEntries, Hostname: runner.Hostname()}).Publish(".dryrun")
	} else {
		cmd := fmt.Sprintf("cd %s && rm -f *.failed %s", checksumDir, strings.Join(invalidCacheEntries, " "))
		m := &pubsub.Message{Key: pubsub.MessageUrknallInternal, Hostname: runner.Hostname()}
		m.Publish("started")

		c, e := runner.Command(cmd)
		if e != nil {
			return e
		}
		if e := c.Run(); e != nil {
			return e
		}
		//m.sshResult = result
		m.Publish("finished")
	}
	return nil
}
