package urknall

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dynport/urknall/cmd"
)

type Runner struct {
	DryRun bool
	Env    []string
	target Target
}

func NewRunner(tgt Target) *Runner {
	return &Runner{target: tgt}
}

func (runner *Runner) Command(cmd string) (cmd.ExecCommand, error) {
	if runner.target.User() != "root" {
		cmd = fmt.Sprintf("sudo sh -c %q", cmd)
	}
	return runner.target.Command(cmd)
}

func (runner *Runner) prepare() error {
	if runner.target.User() == "" {
		return fmt.Errorf("User not set")
	}
	cmd, e := runner.Command(fmt.Sprintf(`{ grep "^%s:" /etc/group | grep %s; } && [[ -d /var/lib/urknall ]]`, ukGROUP, runner.target.User()))
	if e != nil {
		return e
	}
	if e := cmd.Run(); e != nil {
		// If user is missing the group, create group (if necessary), add user and restart ssh connection.
		cmds := []string{
			fmt.Sprintf(`{ grep -e '^%[1]s:' /etc/group > /dev/null || { groupadd %[1]s; }; }`, ukGROUP),
			fmt.Sprintf(`{ [[ -d %[1]s ]] || { mkdir -p -m 2775 %[1]s && chgrp %[2]s %[1]s; }; }`, ukCACHEDIR, ukGROUP),
			fmt.Sprintf("usermod -a -G %s %s", ukGROUP, runner.target.User()),
		}

		cmd, e = runner.Command(fmt.Sprintf(`sudo bash -c "%s"`, strings.Join(cmds, " && ")))
		if e != nil {
			return e
		}
		out := &bytes.Buffer{}
		err := &bytes.Buffer{}
		cmd.SetStderr(err)
		cmd.SetStdout(out)
		if e := cmd.Run(); e != nil {
			return fmt.Errorf("failed to initiate user %q for provisioning: %s, out=%q err=%q", runner.User(), e, out.String(), err.String())
		}
	}
	return nil
}

func (runner *Runner) Hostname() string {
	if s, ok := runner.target.(fmt.Stringer); ok {
		return s.String()
	}
	return "MISSING"
}
