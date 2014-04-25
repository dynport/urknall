package local

import (
	"io"
	"os/exec"

	"github.com/dynport/urknall/runner"
)

type Commander struct {
}

type Command struct {
	command *exec.Cmd
}

func (c *Commander) String() string {
	return "LOCAL"
}

func (c *Commander) Command(cmd string) (runner.Command, error) {
	return &Command{
		command: exec.Command("bash", "-c", cmd),
	}, nil
}

func (c *Command) StdoutPipe() (io.Reader, error) {
	return c.command.StdoutPipe()
}

func (c *Command) StderrPipe() (io.Reader, error) {
	return c.command.StderrPipe()
}

func (c *Command) SetStdout(w io.Writer) {
	c.command.Stdout = w
}

func (c *Command) SetStderr(w io.Writer) {
	c.command.Stderr = w
}

func (c *Command) SetStdin(r io.Reader) {
	c.command.Stdin = r
}

func (c *Command) Run() error {
	return c.command.Run()
}
