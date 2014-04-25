package urknall

import (
	"bytes"
	"fmt"
	"strings"
)

type Runner struct {
	User      string
	DryRun    bool
	Env       []string
	Commander Commander
}

func (runner *Runner) Hostname() string {
	if s, ok := runner.Commander.(fmt.Stringer); ok {
		return s.String()
	}
	return "MISSING"
}

func (runner *Runner) IsSudoRequired() bool {
	return runner.User != "root"
}

func prepareHost(runner *Runner) error {
	if runner.User == "" {
		return fmt.Errorf("User not set")
	}
	cmd, e := runner.Commander.Command(fmt.Sprintf(`{ grep "^%s:" /etc/group | grep %s; } && [[ -d /var/lib/urknall ]]`, ukGROUP, runner.User))
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
