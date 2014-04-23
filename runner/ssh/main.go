package ssh

import (
	"io"

	"code.google.com/p/go.crypto/ssh"
	"github.com/dynport/urknall/runner"
)

type Commander struct {
	Address  string
	User     string
	Password string

	client *ssh.Client
}

func (c *Commander) Command(cmd string) (runner.Command, error) {
	if c.client == nil {
		var e error
		config := &ssh.ClientConfig{
			User: c.User,
		}
		if c.Password != "" {
			config.Auth = append(config.Auth, ssh.Password(c.Password))
		}
		con, e := ssh.Dial("tcp", c.Address, config)
		if e != nil {
			return nil, e
		}
		c.client = &ssh.Client{Conn: con}
	}
	ses, e := c.client.NewSession()
	if e != nil {
		return nil, e
	}
	return &Command{command: cmd, session: ses}, nil
}

type Command struct {
	command string
	session *ssh.Session
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
