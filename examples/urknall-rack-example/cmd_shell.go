package main

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall/utils"
)

func Shell(cmd string) *ShellCommand {
	return &ShellCommand{Command: cmd}
}

type ShellCommand struct {
	Command string // Command to be executed in the shell.
	user    string // User to run the command as.
}

func (cmd *ShellCommand) Render(i interface{}) {
	cmd.Command = utils.MustRenderTemplate(cmd.Command, i)
}

func (sc *ShellCommand) Shell() string {
	if sc.isExecutedAsUser() {
		return fmt.Sprintf("su -l %s <<EOF_ZWO_ASUSER\n%s\nEOF_ZWO_ASUSER\n", sc.user, sc.Command)
	}
	return sc.Command
}

func (sc *ShellCommand) isExecutedAsUser() bool {
	return sc.user != "" && sc.user != "root"
}

func (sc *ShellCommand) Logging() string {
	s := []string{"[COMMAND]"}

	if sc.isExecutedAsUser() {
		s = append(s, fmt.Sprintf("[SU:%s]", sc.user))
	}

	s = append(s, fmt.Sprintf(" # %s", sc.Command))

	return strings.Join(s, "")
}
