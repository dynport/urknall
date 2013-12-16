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
	return &sshClient{host: host, client: gossh.New(host.IP, host.user()), provisionOptions: *opts}
}

func (sc *sshClient) provision() (e error) {
	if e = sc.host.precompileRunlists(); e != nil {
		return e
	}

	return provisionRunlists(sc.host.runlists(), sc.provisionRunlist)
}

func (sc *sshClient) provisionRunlist(rl *Runlist) (e error) {
	tasks := sc.buildTasksForRunlist(rl)

	checksumDir := fmt.Sprintf("/var/cache/urknall/%s", rl.name)

	checksumHash, e := sc.buildChecksumHash(checksumDir)
	if e != nil {
		return fmt.Errorf("failed to build checksum hash: %s", e.Error())
	}

	for i := range tasks {
		started := time.Now()
		task := tasks[i]
		logMsg := task.command.Logging()
		if _, found := checksumHash[task.checksum]; found { // Task is cached.
			Publish("runlists.provision.task.cached", &Message{ExecStatus: StatusCached, Task: task, Message: logMsg, Host: sc.host, Runlist: rl,
				Duration: time.Now().Sub(started),
			})
			delete(checksumHash, task.checksum) // Delete checksums of cached tasks from hash.
			continue
		}

		if len(checksumHash) > 0 { // All remaining checksums are invalid, as something changed.
			if e = sc.cleanUpRemainingCachedEntries(checksumDir, checksumHash); e != nil {
				return e
			}
			checksumHash = make(map[string]struct{})
		}

		Publish("runlists.provision.task.start", &Message{ExecStatus: StatusExecStart, Task: task, Message: logMsg, Host: sc.host, Runlist: rl})
		e = sc.runTask(task, checksumDir)
		Publish("runlists.provision.task.finished", &Message{
			ExecStatus: StatusExecFinished, Task: task, Message: logMsg, Host: sc.host, Duration: time.Now().Sub(started),
			Runlist: rl,
			Error:   e,
		},
		)
		if e != nil {
			return e
		}
	}

	return nil
}

func newDebugWriter(stream string, host *Host, task *taskData) func(i ...interface{}) {
	started := time.Now()
	return func(i ...interface{}) {
		var runlist *Runlist = nil
		if task != nil {
			runlist = task.runlist
		}
		Publish("task.io."+stream, &Message{Host: host, Stream: stream, Task: task, IOMessage: i, Runlist: runlist, Duration: time.Now().Sub(started)})
	}
}

func (sc *sshClient) runTask(task *taskData, checksumDir string) (e error) {
	if sc.provisionOptions.DryRun {
		return nil
	}

	stderr := fmt.Sprintf(`>(while read line; do echo "$(date --iso-8601=ns):stderr:$line"; done | tee /tmp/%s.%s.stderr)`, sc.host.user(), task.checksum)
	stdout := fmt.Sprintf(`>(while read line; do echo "$(date --iso-8601=ns):stdout:$line"; done | tee /tmp/%s.%s.stdout)`, sc.host.user(), task.checksum)

	sc.client.DebugWriter = newDebugWriter("stdout", sc.host, task)
	sc.client.ErrorWriter = newDebugWriter("stderr", sc.host, task)

	sCmd := fmt.Sprintf("bash <<EOF_RUNTASK 1> %s 2> %s\n%s\nEOF_RUNTASK\n", stdout, stderr, task.command.Shell())
	if sc.host.isSudoRequired() {
		sCmd = fmt.Sprintf("sudo %s", sCmd)
	}
	rsp, e := sc.client.Execute(sCmd)

	// Write the checksum file (containing information on the command run).
	sc.writeChecksumFile(checksumDir, task.checksum, e != nil, task.command.Logging(), rsp)

	if e != nil {
		return fmt.Errorf("%s (see %s/%s.failed for more information)", e.Error(), checksumDir, task.checksum)
	}
	return nil
}

func (sc *sshClient) executeCommand(cmdRaw string) *gossh.Result {
	cmdRaw = fmt.Sprintf("bash <<EOF_ZWO_SUDO\n%s\nEOF_ZWO_SUDO\n", cmdRaw)
	if sc.host.isSudoRequired() {
		cmdRaw = "sudo " + cmdRaw
	}
	c := &cmd.ShellCommand{Command: cmdRaw}
	result, e := sc.client.Execute(c.Shell())
	if e != nil {
		stderr := ""
		if result != nil {
			stderr = strings.TrimSpace(result.Stderr())
		}
		panic(fmt.Errorf("internal error: %s (%s)", e.Error(), stderr))
	}
	return result
}

func (sc *sshClient) buildChecksumHash(checksumDir string) (checksumMap map[string]struct{}, e error) {
	// Make sure the directory exists.
	sc.executeCommand(fmt.Sprintf("mkdir -p %s", checksumDir))

	checksums := []string{}
	// The subshell for the if state requires the escaping of the '$' so that the variable is only expanded in the
	// subshell.
	rsp := sc.executeCommand(fmt.Sprintf(`for f in "%s"/*.done; do if [[ -f "\$f" ]]; then echo -n "\$f "; fi; done`, checksumDir))
	for _, checksumFile := range strings.Fields(rsp.Stdout()) {
		checksum := strings.TrimSuffix(path.Base(checksumFile), ".done")
		checksums = append(checksums, checksum)
	}

	checksumMap = make(map[string]struct{})
	for i := range checksums {
		if len(checksums[i]) != 64 {
			return nil, fmt.Errorf("invalid checksum '%s' found in '%s'", checksums[i], checksumDir)
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
		Publish("urknall.cleanup_cache_entries", &Message{InvalidatedCachentries: invalidCacheEntries, Host: sc.host})
	} else {
		cmd := fmt.Sprintf("cd %s && rm -f *.failed %s", checksumDir, strings.Join(invalidCacheEntries, " "))
		Publish("urknall.cleanup_cache_entries", &Message{Command: cmd, Host: sc.host})
		sc.executeCommand(cmd)
	}
	return nil
}

func (sc *sshClient) writeChecksumFile(checksumDir, checksum string, failed bool, logMsg string, response *gossh.Result) {
	tmpChecksumFiles := "/tmp/" + sc.host.user() + "." + checksum + ".std*"
	checksumFile := checksumDir + "/" + checksum
	if failed {
		checksumFile += ".failed"
	} else {
		checksumFile += ".done"
	}

	// Whoa, super hacky stuff to get the command to the checksum file. The command might contain a lot of stuff, like
	// apostrophes and the like, that would totally nuke a quoted string. Though there is a here doc.
	c := []string{
		fmt.Sprintf(`cat %s | sort >> %s`, tmpChecksumFiles, checksumFile),
		fmt.Sprintf(`rm -f %s`, tmpChecksumFiles),
	}
	sc.executeCommand(fmt.Sprintf("cat <<EOF_COMMAND > %s && %s\n%s\nEOF_COMMAND\n", checksumFile, strings.Join(c, " && "), logMsg))
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
