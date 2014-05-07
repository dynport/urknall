package urknall

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dynport/urknall/cmd"
)

func (build *Build) prepareCommand(cmd string) (cmd.ExecCommand, error) {
	if build.User() != "root" {
		cmd = fmt.Sprintf("sudo sh -c %q", cmd)
	}
	return build.Command(cmd)
}

func (build *Build) prepare() error {
	if build.User() == "" {
		return fmt.Errorf("User not set")
	}
	cmd, e := build.prepareCommand(fmt.Sprintf(`{ grep "^%s:" /etc/group | grep %s; } && [ -d /var/lib/urknall ]`, ukGROUP, build.User()))
	if e != nil {
		return e
	}
	if e := cmd.Run(); e != nil {
		// If user is missing the group, create group (if necessary), add user and restart ssh connection.
		cmds := []string{
			fmt.Sprintf(`{ grep -e '^%[1]s:' /etc/group > /dev/null || { groupadd %[1]s; }; }`, ukGROUP),
			fmt.Sprintf(`{ [ -d %[1]s ] || { mkdir -p -m 2775 %[1]s && chgrp %[2]s %[1]s; }; }`, ukCACHEDIR, ukGROUP),
			fmt.Sprintf("usermod -a -G %s %s", ukGROUP, build.User()),
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
	}
	return nil
}

func (build *Build) hostname() string {
	if s, ok := build.Target.(fmt.Stringer); ok {
		return s.String()
	}
	return "MISSING"
}
