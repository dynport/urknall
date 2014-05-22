package urknall

import "github.com/dynport/urknall/cmd"

type rawCommand struct {
	cmd.Command // The command to be executed.
	taskName    string
}

func (c *rawCommand) TaskName() string {
	return c.taskName
}
