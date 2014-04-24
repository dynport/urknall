package urknall

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dynport/gossh"
	"github.com/dynport/urknall/cmd"
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
	c.Port = host.Port
	if host.Password != "" {
		c.SetPassword(host.Password)
	}
	return &sshClient{host: host, client: c, provisionOptions: *opts}
}

func buildBinaryPackage(runner *Runner, pkg BinaryPackage) (e error) {
	name := pkg.Name() + "." + pkg.PkgVersion()
	compileRunlist := newRunlist(name+".build", pkg, nil)
	// Don't use binary packages, as otherwise you'll be caugth in an awkward self reference trying to use a binary
	// package to build the very binary package.
	if e = compileRunlist.compileWithoutBinaryPackages(); e != nil {
		return e
	}

	packageRunlist := newRunlist(name+".package", pkg, nil)
	if e = packageRunlist.buildBinaryPackage(); e != nil {
		return e
	}

	return provisionRunlists([]*Package{compileRunlist, packageRunlist}, runner)
}

func prepareHost(runner *Runner) error {
	if runner.User == "" {
		return fmt.Errorf("User not set")
	}
	cmd, e := runner.Commander.Command(fmt.Sprintf(`grep "^%s:" /etc/group | grep %s`, ukGROUP, runner.User))
	if e != nil {
		return e
	}
	if e := cmd.Run(); e != nil {
		// If user is missing the group, create group (if necessary), add user and restart ssh connection.
		cmds := []string{
			fmt.Sprintf(`{ grep -e '^%[1]s:' /etc/group > /dev/null || { groupadd %[1]s; }; }`, ukGROUP),
			fmt.Sprintf(`{ [[ -d %[1]s ]] || { mkdir -p -m 2775 %[1]s && chgrp %[2]s %[1]s; }; }`, ukCACHEDIR, ukGROUP),
			fmt.Sprintf("usermod -a -G %s %s", ukGROUP, runner.User),
		}

		cmd, e = runner.Commander.Command(fmt.Sprintf(`sudo bash -c "%s"`, strings.Join(cmds, " && ")))
		if e != nil {
			return e
		}
		out := &bytes.Buffer{}
		err := &bytes.Buffer{}
		cmd.SetStderr(err)
		cmd.SetStdout(out)
		if e := cmd.Run(); e != nil {
			return fmt.Errorf("failed to initiate user %q for provisioning: %s, out=%q err=%q", runner.User, e, out.String(), err.String())
		}
	}
	return nil
}

func provisionRunlist(runner *Runner, rl *Package, ct checksumTree) (e error) {
	tasks := rl.tasks()

	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", rl.name)

	var found bool
	var checksumHash map[string]struct{}
	if checksumHash, found = ct[rl.name]; !found {
		ct[rl.name] = map[string]struct{}{}
		checksumHash = ct[rl.name]

		// Create checksum dir and set group bit (all new files will inherit the directory's group). This allows for
		// different users (being part of that group) to create, modify and delete the contained checksum and log files.
		createChecksumDirCmd := fmt.Sprintf("mkdir -m2775 -p %s", checksumDir)
		if runner.IsSudoRequired() {
			createChecksumDirCmd = fmt.Sprintf(`sudo %s`, createChecksumDirCmd)
		}

		cmd, e := runner.Commander.Command(createChecksumDirCmd)
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
		m := &Message{key: MessageRunlistsProvisionTask, task: task, message: logMsg, host: nil, runlist: rl}
		if _, found := checksumHash[task.checksum]; found { // Task is cached.
			m.execStatus = statusCached
			m.publish("finished")
			delete(checksumHash, task.checksum) // Delete checksums of cached tasks from hash.
			continue
		}

		if len(checksumHash) > 0 { // All remaining checksums are invalid, as something changed.
			if e = cleanUpRemainingCachedEntries(runner, checksumDir, checksumHash); e != nil {
				return e
			}
			checksumHash = make(map[string]struct{})
		}
		m.execStatus = statusExecStart
		m.publish("started")
		e = runTask(runner, task, checksumDir)
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
			var runlist *Package = nil
			if task != nil {
				runlist = task.runlist
			}
			m := &Message{key: "task.io", host: host, stream: stream, task: task, line: line, runlist: runlist, totalRuntime: time.Now().Sub(started)}
			m.publish(stream)
		}
	}
}

func runTask(runner *Runner, task *taskData, checksumDir string) (e error) {
	if runner.DryRun {
		return nil
	}
	sCmd := fmt.Sprintf("bash <<EOF_RUNTASK\nset -xe\n%s\nEOF_RUNTASK\n", task.command.Shell())
	for _, env := range runner.Env {
		sCmd = env + " " + sCmd
	}
	if runner.IsSudoRequired() {
		sCmd = fmt.Sprintf("sudo -i %s", sCmd)
	}
	r := &remoteTaskRunner{Runner: runner, cmd: sCmd, task: task, dir: checksumDir}
	return r.run()
}

func buildChecksumTree(runner *Runner) (ct checksumTree, e error) {
	ct = checksumTree{}

	cmd, e := runner.Commander.Command(fmt.Sprintf(`[[ -d %[1]s ]] && find %[1]s -type f -name \*.done`, ukCACHEDIR))
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

func cleanUpRemainingCachedEntries(runner *Runner, checksumDir string, checksumHash map[string]struct{}) (e error) {
	invalidCacheEntries := make([]string, 0, len(checksumHash))
	for k, _ := range checksumHash {
		invalidCacheEntries = append(invalidCacheEntries, fmt.Sprintf("%s.done", k))
	}
	if runner.DryRun {
		(&Message{key: MessageCleanupCacheEntries, invalidatedCachentries: invalidCacheEntries, host: nil}).publish(".dryrun")
	} else {
		cmd := fmt.Sprintf("cd %s && rm -f *.failed %s", checksumDir, strings.Join(invalidCacheEntries, " "))
		m := &Message{command: cmd, host: nil, key: MessageUrknallInternal}
		m.publish("started")

		c, e := runner.Commander.Command(cmd)
		if e != nil {
			return e
		}
		if e := c.Run(); e != nil {
			return e
		}
		//m.sshResult = result
		m.publish("finished")
	}
	return nil
}

type taskData struct {
	command  cmd.Command // The command to be executed.
	checksum string      // The checksum of the command.
	runlist  *Package
}

func (data *taskData) Command() cmd.Command {
	return data.command
}
