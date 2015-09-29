package urknall

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dynport/urknall/pubsub"
	"github.com/dynport/urknall/target"
)

// A shortcut creating and running a build from the given target and template.
func Run(target Target, tpl Template) (e error) {
	return (&Build{Target: target, Template: tpl}).Run()
}

// A shortcut creating and runnign a build from the given target and template
// with the DryRun flag set to true. This is quite helpful to actually see
// which commands would be exeucted in the current setting, without actually
// doing anything.
func DryRun(target Target, tpl Template) (e error) {
	return (&Build{Target: target, Template: tpl}).DryRun()
}

// A build is the glue between a target and template.
type Build struct {
	Target            // Where to run the build.
	Template          // What to actually build.
	Env      []string // Environment variables in the form `KEY=VALUE`.
}

// This will render the build's template into a package and run all its tasks.
func (b *Build) Run() error {
	pkg, e := b.prepareBuild()
	if e != nil {
		return e
	}
	m := message(pubsub.MessageTasksProvision, b.hostname(), "")
	m.Publish("started")
	for _, task := range pkg.tasks {
		if e = b.buildTask(task); e != nil {
			m.PublishError(e)
			return e
		}
	}
	m.Publish("finished")
	return nil
}

func (b *Build) DryRun() error {
	pkg, e := b.prepareBuild()
	if e != nil {
		return e
	}

	for _, task := range pkg.tasks {
		for _, command := range task.commands {
			m := message(pubsub.MessageTasksProvisionTask, b.hostname(), task.name)
			m.TaskChecksum = command.Checksum()
			m.Message = command.LogMsg()

			switch {
			case command.cached:
				m.ExecStatus = pubsub.StatusCached
				m.Publish("finished")
			default:
				m.ExecStatus = pubsub.StatusExecStart
				m.Publish("executed")
			}
		}
	}
	return nil
}

func (build *Build) prepareBuild() (*packageImpl, error) {
	pkg, e := renderTemplate(build.Template)
	if e != nil {
		return nil, e
	}

	if e = build.prepareTarget(); e != nil {
		return nil, e
	}

	ct, e := build.buildChecksumTree()
	if e != nil {
		return nil, fmt.Errorf("error building checksum tree: %s", e.Error())
	}

	for _, task := range pkg.tasks {
		if e = build.prepareTask(task, ct); e != nil {
			return nil, e
		}
	}

	return pkg, nil
}

func (build *Build) prepareTarget() error {
	if build.User() == "" {
		return fmt.Errorf("User not set")
	}
	rawCmd := fmt.Sprintf(`{ grep "^%s:" /etc/group | grep %s; } && [ -d %[3]s ] && [ -f %[3]s/.v2 ]`,
		ukGROUP, build.User(), ukCACHEDIR)
	cmd, e := build.prepareInternalCommand(rawCmd)
	if e != nil {
		return e
	}
	if e := cmd.Run(); e != nil {
		// If user is missing the group, create group (if necessary), add user and restart ssh connection.
		cmds := []string{
			fmt.Sprintf(`{ grep -e '^%[1]s:' /etc/group > /dev/null || { groupadd %[1]s; }; }`, ukGROUP),
			fmt.Sprintf(`{ [ -d %[1]s ] || { mkdir -p -m 2775 %[1]s && chgrp %[2]s %[1]s; }; }`, ukCACHEDIR, ukGROUP),
			fmt.Sprintf("usermod -a -G %s %s", ukGROUP, build.User()),
			fmt.Sprintf(`[ -f %[1]s/.v2 ] || { export DATE=$(date "+%%Y%%m%%d_%%H%%M%%S") && ls %[1]s | while read dir; do ls -t %[1]s/$dir/*.done | tac > %[1]s/$dir/$DATE.run; done && touch %[1]s/.v2;  } `, ukCACHEDIR),
		}

		cmd, e = build.prepareInternalCommand(strings.Join(cmds, " && "))
		if e != nil {
			return e
		}
		out := &bytes.Buffer{}
		err := &bytes.Buffer{}
		cmd.SetStderr(err)
		cmd.SetStdout(out)
		if e := cmd.Run(); e != nil {
			return fmt.Errorf("failed to initiate user %q for provisioning: %s, out=%q err=%q", build.User(), e, out.String(), err.String())
		}
		return build.Reset()
	}
	return nil
}

func (build *Build) prepareTask(tsk *task, ct checksumTree) (e error) {
	cacheKey := tsk.name
	if cacheKey == "" {
		return fmt.Errorf("CacheKey must not be empty")
	}
	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", tsk.name)

	var found bool
	var checksumList []string

	if checksumList, found = ct[cacheKey]; !found {
		// Create checksum dir and set group bit (all new files will inherit the directory's group). This allows for
		// different users (being part of that group) to create, modify and delete the contained checksum and log files.
		createChecksumDirCmd := fmt.Sprintf("mkdir -m2775 -p %s", checksumDir)

		cmd, e := build.prepareInternalCommand(createChecksumDirCmd)
		if e != nil {
			return e
		}
		err := &bytes.Buffer{}

		cmd.SetStderr(err)

		if e := cmd.Run(); e != nil {
			return fmt.Errorf(err.String() + ": " + e.Error())
		}
	}

	// find commands that need not be executed
	for i, cmd := range tsk.commands {
		checksum, e := commandChecksum(cmd.command)
		if e != nil {
			return e
		}

		switch {
		case len(checksumList) <= i || checksum != checksumList[i]:
			return nil
		default:
			cmd.cached = true
		}
	}

	return nil
}

func (build *Build) buildTask(tsk *task) (e error) {
	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", tsk.name)

	tsk.started = time.Now()

	for _, cmd := range tsk.commands {
		checksum := cmd.Checksum()

		m := message(pubsub.MessageTasksProvisionTask, build.hostname(), tsk.name)
		m.TaskChecksum = checksum
		m.Message = cmd.LogMsg()

		var cmdErr error

		switch {
		case cmd.cached:
			m.ExecStatus = pubsub.StatusCached
		default:
			m.ExecStatus = pubsub.StatusExecStart
			m.Publish("started")

			r := &commandRunner{
				build:       build,
				command:     cmd.command,
				dir:         checksumDir,
				taskName:    tsk.name,
			}
			cmdErr = r.run()

			m.Error = cmdErr
			m.ExecStatus = pubsub.StatusExecFinished
		}
		m.Publish("finished")

		err := build.addCmdToTaskLog(tsk, checksumDir, checksum, cmdErr)
		switch {
		case cmdErr != nil:
			return cmdErr
		case err != nil:
			return err
		}
	}

	return nil
}

// addCmdToTaskLog will manage the log of run commands in a file. This file gets append the path to a file
// for each command, that contains the executed script. The filename contains either ".done" or ".failed" as
// suffix, depending on the err given (nil or not).
func (build *Build) addCmdToTaskLog(tsk *task, checksumDir, checksum string, err error) (e error) {
	prefix := checksumDir + "/" + checksum
	sourceFile := prefix + ".sh"
	targetFile := prefix + ".done"
	if err != nil {
		logError(err)
		targetFile = prefix + ".failed"
	}
	rawCmd := fmt.Sprintf("{ [ -f %[1]s ] || mv %[2]s %[1]s; } && echo %[1]s >> %[3]s/%[4]s.run",
		targetFile, sourceFile, checksumDir, tsk.started.Format("20060102_150405"))
	c, e := build.prepareInternalCommand(rawCmd)
	if e != nil {
		return e
	}

	return c.Run()
}

type checksumTree map[string][]string

func (build *Build) buildChecksumTree() (ct checksumTree, e error) {
	ct = checksumTree{}

	rawCmd := fmt.Sprintf(
		`[ -d %[1]s ] && { ls %[1]s | while read dir; do ls -t %[1]s/$dir/*.run | head -n1 | xargs cat; done; }`,
		ukCACHEDIR)
	cmd, e := build.prepareInternalCommand(rawCmd)
	if e != nil {
		return nil, e
	}
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}

	cmd.SetStdout(out)
	cmd.SetStderr(err)

	if e := cmd.Run(); e != nil {
		return nil, fmt.Errorf("%s: out=%s err=%s", e.Error(), out.String(), err.String())
	}

	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)

		if line == "" || !strings.HasSuffix(line, ".done") {
			continue
		}

		pkgname := filepath.Dir(strings.TrimPrefix(line, ukCACHEDIR+"/"))
		checksum := strings.TrimSuffix(filepath.Base(line), ".done")
		if len(checksum) != 64 {
			return nil, fmt.Errorf("invalid checksum %q found for package %q", checksum, pkgname)
		}
		ct[pkgname] = append(ct[pkgname], checksum)
	}

	return ct, nil
}

func (build *Build) prepareCommand(rawCmd string) (target.ExecCommand, error) {
	var sudo string
	if build.User() != "root" {
		sudo = "sudo "
	}
	return build.Command(sudo + rawCmd)
}

func (build *Build) prepareInternalCommand(rawCmd string) (target.ExecCommand, error) {
	rawCmd = fmt.Sprintf("sh -x -e <<\"EOC\"\n%s\nEOC\n", rawCmd)
	return build.prepareCommand(rawCmd)
}

func (build *Build) hostname() string {
	if s, ok := build.Target.(fmt.Stringer); ok {
		return s.String()
	}
	return "MISSING"
}
