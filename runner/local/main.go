package local

import (
	"io"
	"os/exec"

	"github.com/dynport/urknall/runner"
)

type Host struct {
}

func (c *Host) String() string {
	return "LOCAL"
}

func (c *Host) Command(cmd string) (runner.Command, error) {
	return &Command{
		command: exec.Command("bash", "-c", cmd),
	}, nil
}

type Command struct {
	command *exec.Cmd
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
