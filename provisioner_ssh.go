package urknall

import (
	"crypto/sha256"
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

func (sc *sshClient) buildBinaryPackage(pkg BinaryPackage) (e error) {
	name := pkg.Name() + "." + pkg.PkgVersion()
	compileRunlist := newRunlist(name+".build", pkg, sc.host)
	// Don't use binary packages, as otherwise you'll be caugth in an awkward self reference trying to use a binary
	// package to build the very binary package.
	if e = compileRunlist.compileWithoutBinaryPackages(); e != nil {
		return e
	}

	packageRunlist := newRunlist(name+".package", pkg, sc.host)
	if e = packageRunlist.buildBinaryPackage(); e != nil {
		return e
	}

	return provisionRunlists([]*Package{compileRunlist, packageRunlist}, sc)
}

func (sc *sshClient) provision() (e error) {
	if e = sc.host.precompileRunlists(); e != nil {
		return e
	}

	if e = sc.prepareHost(); e != nil {
		return e
	}

	return provisionRunlists(sc.host.runlists(), sc)
}

func (sc *sshClient) prepareHost() (e error) {
	con, e := sc.client.Connection()
	if e != nil {
		return e
	}

	if e := executeCommand(con, fmt.Sprintf(`grep "^%s:" /etc/group | grep %s`, ukGROUP, sc.host.user())); e != nil {
		// If user is missing the group, create group (if necessary), add user and restart ssh connection.
		cmds := []string{
			fmt.Sprintf(`{ grep -e '^%[1]s:' /etc/group > /dev/null || { groupadd %[1]s; }; }`, ukGROUP),
			fmt.Sprintf(`{ [[ -d %[1]s ]] || { mkdir -p -m 2775 %[1]s && chgrp %[2]s %[1]s; }; }`, ukCACHEDIR, ukGROUP),
			fmt.Sprintf("usermod -a -G %s %s", ukGROUP, sc.host.user()),
		}

		if e := executeCommand(con, fmt.Sprintf(`sudo bash -c "%s"`, strings.Join(cmds, " && "))); e != nil {
			return fmt.Errorf("failed to initiate user %q for provisioning: %s", sc.host.user(), e)
		}

		// Restarting the connection is required to make sure the user's new group is added properly.
		sc.client.Conn.Close()
		sc.client.Conn = nil
	}
	return nil
}

func (sc *sshClient) ProvisionRunlist(rl *Package, ct checksumTree) (e error) {
	tasks := sc.buildTasksForRunlist(rl)

	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", rl.name)

	var found bool
	var checksumHash map[string]struct{}
	if checksumHash, found = ct[rl.name]; !found {
		ct[rl.name] = map[string]struct{}{}
		checksumHash = ct[rl.name]

		// Create checksum dir and set group bit (all new files will inherit the directory's group). This allows for
		// different users (being part of that group) to create, modify and delete the contained checksum and log files.
		createChecksumDirCmd := fmt.Sprintf("mkdir -m2775 -p %s", checksumDir)
		if sc.host.isSudoRequired() {
			createChecksumDirCmd = fmt.Sprintf(`sudo %s`, createChecksumDirCmd)
		}
		r, e := sc.client.Execute(createChecksumDirCmd)
		if e != nil {
			return fmt.Errorf(r.Stderr() + ": " + e.Error())
		}
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
			var runlist *Package = nil
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
	for _, env := range sc.host.Env {
		sCmd = env + " " + sCmd
	}
	if sc.host.isSudoRequired() {
		sCmd = fmt.Sprintf("sudo -i %s", sCmd)
	}

	con, e := sc.client.Connection()
	if e != nil {
		return e
	}
	runner := &remoteTaskRunner{clientConn: con, cmd: sCmd, task: task, host: sc.host, dir: checksumDir}
	return runner.run()
}

func (sc *sshClient) BuildChecksumTree() (ct checksumTree, e error) {
	ct = checksumTree{}

	rsp, e := sc.client.Execute(fmt.Sprintf(`[[ -d %[1]s ]] && find %[1]s -type f -name \*.done`, ukCACHEDIR))
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
	runlist  *Package
}

func (data *taskData) Command() cmd.Command {
	return data.command
}

func (sc *sshClient) buildTasksForRunlist(rl *Package) (tasks []*taskData) {
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
