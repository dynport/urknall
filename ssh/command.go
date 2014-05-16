package ssh

import (
	"io"

	"code.google.com/p/go.crypto/ssh"
)

type sshCommand struct {
	command string
	session *ssh.Session
}

func (c *sshCommand) Close() error {
	return c.session.Close()
}

func (c *sshCommand) StdinPipe() (io.Writer, error) {
	return c.session.StdinPipe()
}

func (c *sshCommand) StdoutPipe() (io.Reader, error) {
	return c.session.StdoutPipe()
}

func (c *sshCommand) StderrPipe() (io.Reader, error) {
	return c.session.StderrPipe()
}

func (c *sshCommand) SetStdout(w io.Writer) {
	c.session.Stdout = w
}

func (c *sshCommand) SetStderr(w io.Writer) {
	c.session.Stderr = w
}

func (c *sshCommand) SetStdin(r io.Reader) {
	c.session.Stdin = r
}

func (c *sshCommand) Run() error {
	return c.session.Run(c.command)
}

func (c *sshCommand) Wait() error {
	return c.session.Wait()
}

func (c *sshCommand) Start() error {
	return c.session.Start(c.command)
}
