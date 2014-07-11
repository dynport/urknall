package main

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall/utils"
)

// A shell command is just that: something that is executed in a shell on the host to be provisioned. There is quite a
// lot of infrastructure to build such commands. To make construction of complicated commands easier those helpers use
// the most generic type "interface{}". Thereby it is possible to use these functions with "strings" or other
// "ShellCommands" (returned by other helpers for example).
//
// There are some commands that relate to the system's package management. Those are currently based on apt, i.e. only
// debian based systems can be used (our current system of choice is ubuntu server in version 12.04LTS as of this
// writing).
type ShellCommand struct {
	Command string // Command to be executed in the shell.
	user    string // User to run the command as.
}

func (cmd *ShellCommand) Render(i interface{}) {
	cmd.Command = utils.MustRenderTemplate(cmd.Command, i)
	if cmd.user != "" {
		cmd.user = utils.MustRenderTemplate(cmd.user, i)
	}
}

func Shell(cmd string) *ShellCommand {
	return &ShellCommand{Command: cmd}
}

func (sc *ShellCommand) Shell() string {
	if sc.isExecutedAsUser() {
		return fmt.Sprintf("su -l %s <<EOF_ZWO_ASUSER\n%s\nEOF_ZWO_ASUSER\n", sc.user, sc.Command)
	}
	return sc.Command
}

func (sc *ShellCommand) Logging() string {
	s := []string{"[COMMAND]"}

	if sc.isExecutedAsUser() {
		s = append(s, fmt.Sprintf("[SU:%s]", sc.user))
	}

	s = append(s, fmt.Sprintf(" # %s", sc.Command))

	return strings.Join(s, "")
}

func (sc *ShellCommand) isExecutedAsUser() bool {
	return sc.user != "" && sc.user != "root"
}
