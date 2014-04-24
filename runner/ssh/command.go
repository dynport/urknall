package ssh

import (
	"io"

	"code.google.com/p/go.crypto/ssh"
)

type Command struct {
	command string
	session *ssh.Session
}

func (c *Command) StdinPipe() (io.Writer, error) {
	return c.session.StdinPipe()
}

func (c *Command) StdoutPipe() (io.Reader, error) {
	return c.session.StdoutPipe()
}

func (c *Command) StderrPipe() (io.Reader, error) {
	return c.session.StderrPipe()
}

func (c *Command) SetStdout(w io.Writer) {
	c.session.Stdout = w
}

func (c *Command) SetStderr(w io.Writer) {
	c.session.Stderr = w
}

func (c *Command) SetStdin(r io.Reader) {
	c.session.Stdin = r
}

func (c *Command) Run() error {
	return c.session.Run(c.command)
}

func (c *Command) Wait() error {
	return c.session.Wait()
}

func (c *Command) Start() error {
	return c.session.Start(c.command)
}
