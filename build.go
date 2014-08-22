package urknall

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pubsub"
)

// A shortcut for rendering the given template to the given target.
func Run(target Target, tpl Template) (e error) {
	return (&Build{Target: target, Template: tpl}).Run()
}

// A shortcut for rendering the given template to the given target, without
// actually executing any commands. This is quite helpful to see which commands
// would be exeucted in the target's current state.
func DryRun(target Target, tpl Template) (e error) {
	return (&Build{Target: target, Template: tpl}).DryRun()
}

// A build is the glue between a target and template. It contains the basic
// parameters required for actually doing something.
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
	for _, task := range pkg.tasks {
		m := &pubsub.Message{Key: pubsub.MessageRunlistsProvision, Hostname: b.hostname()}
		m.Publish("started")
		if e = b.buildTask(task); e != nil {
			m.PublishError(e)
			return e
		}
		m.Publish("finished")
	}
	return nil
}

func (b *Build) DryRun() error {
	pkg, e := b.prepareBuild()
	if e != nil {
		return e
	}

	for _, task := range pkg.tasks {
		for _, command := range task.commands {
			m := &pubsub.Message{
				Key:          pubsub.MessageRunlistsProvisionTask,
				TaskChecksum: command.Checksum(),
				Message:      command.LogMsg(),
				Hostname:     b.hostname(),
				RunlistName:  task.name,
			}

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
	cmd, e := build.prepareCommand(fmt.Sprintf(`{ grep "^%s:" /etc/group | grep %s; } && [ -d %[3]s ] && [ -f %[3]s/.v2 ]`, ukGROUP, build.User(), ukCACHEDIR))
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

		cmd, e = build.prepareCommand(strings.Join(cmds, " && "))
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

		cmd, e := build.prepareCommand(createChecksumDirCmd)
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
			break
		default:
			cmd.cached = true
		}
	}

	return nil
}

func (build *Build) buildTask(tsk *task) (e error) {
	checksumDir := fmt.Sprintf(ukCACHEDIR+"/%s", tsk.name)

	for _, cmd := range tsk.commands {
		m := &pubsub.Message{
			Key:          pubsub.MessageRunlistsProvisionTask,
			TaskChecksum: cmd.Checksum(),
			Message:      cmd.LogMsg(),
			Hostname:     build.hostname(),
			RunlistName:  tsk.name,
		}

		if cmd.cached { // Task is cached.
			m.ExecStatus = pubsub.StatusCached
			m.Publish("finished")
			continue
		}

		m.ExecStatus = pubsub.StatusExecStart
		m.Publish("started")
		e := executeCommand(cmd.command, build, checksumDir, tsk.name)

		m.Error = e
		m.ExecStatus = pubsub.StatusExecFinished
		m.Publish("finished")

		if e != nil {
			return e
		}
	}

	return nil
}

type checksumTree map[string][]string

func (build *Build) buildChecksumTree() (ct checksumTree, e error) {
	ct = checksumTree{}

	cmd, e := build.prepareCommand(
		fmt.Sprintf(
			`[ -d %[1]s ] && { ls %[1]s | while read dir; do ls -t %[1]s/$dir/*.run | head -n1 | xargs cat; done; }`,
			ukCACHEDIR))
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

func (build *Build) prepareCommand(cmd string) (cmd.ExecCommand, error) {
	var sudo string
	if build.User() != "root" {
		sudo = "sudo "
	}
	cmd = fmt.Sprintf(sudo+"sh -x -e <<\"EOC\"\n%s\nEOC\n", cmd)
	return build.Command(cmd)
}

func (build *Build) hostname() string {
	if s, ok := build.Target.(fmt.Stringer); ok {
		return s.String()
	}
	return "MISSING"
}
