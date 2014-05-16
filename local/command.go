package local

import (
	"io"
	"os/exec"
)

type localCommand struct {
	command *exec.Cmd
}

func (c *localCommand) StdoutPipe() (io.Reader, error) {
	return c.command.StdoutPipe()
}

func (c *localCommand) StderrPipe() (io.Reader, error) {
	return c.command.StderrPipe()
}

func (c *localCommand) StdinPipe() (io.Writer, error) {
	return c.command.StdinPipe()
}

func (c *localCommand) SetStdout(w io.Writer) {
	c.command.Stdout = w
}

func (c *localCommand) SetStderr(w io.Writer) {
	c.command.Stderr = w
}

func (c *localCommand) SetStdin(r io.Reader) {
	c.command.Stdin = r
}

func (c *localCommand) Wait() error {
	return c.command.Wait()
}

func (c *localCommand) Start() error {
	return c.command.Start()
}

func (c *localCommand) Run() error {
	return c.command.Run()
}
