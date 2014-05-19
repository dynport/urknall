package target

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"github.com/dynport/urknall/cmd"
)

func NewSshTarget(addr string) (target *sshTarget, e error) {
	target = &sshTarget{port: 22, user: "root"}

	hostAndPort := strings.SplitN(addr, ":", 2)
	if len(hostAndPort) == 2 {
		addr = hostAndPort[0]
		target.port, e = strconv.Atoi(hostAndPort[1])
		if e != nil {
			return nil, fmt.Errorf("port must be given as integer, got %q", hostAndPort[1])
		}
	}

	userAndAddress := strings.Split(addr, "@")
	switch len(userAndAddress) {
	case 1:
		target.address = addr
	case 2:
		target.user = userAndAddress[0]
		target.address = userAndAddress[1]
	default:
		return nil, fmt.Errorf("expected target address of the form '<user>@<host>', but was given: %s", addr)
	}

	if target.address == "" {
		e = fmt.Errorf("empty address given for target")
	}

	return target, e
}

type sshTarget struct {
	Password string

	user    string
	port    int
	address string

	client *ssh.Client
}

func (target *sshTarget) User() string {
	return target.user
}

func (target *sshTarget) String() string {
	return fmt.Sprintf("%s@%s:%d", target.user, target.address, target.port)
}

func (target *sshTarget) Command(cmd string) (cmd.ExecCommand, error) {
	if target.client == nil {
		var e error
		target.client, e = target.buildClient()
		if e != nil {
			return nil, e
		}
	}
	ses, e := target.client.NewSession()
	if e != nil {
		return nil, e
	}
	return &sshCommand{command: cmd, session: ses}, nil
}

func (target *sshTarget) Reset() (e error) {
	if target.client != nil {
		e = target.client.Close()
		target.client = nil
	}
	return e
}

func (target *sshTarget) buildClient() (*ssh.Client, error) {
	var e error
	config := &ssh.ClientConfig{
		User: target.user,
	}
	if target.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(target.Password))
	} else if sshSocket := os.Getenv("SSH_AUTH_SOCK"); sshSocket != "" {
		if agentConn, e := net.Dial("unix", sshSocket); e == nil {
			config.Auth = append(config.Auth, ssh.PublicKeysCallback(agent.NewClient(agentConn).Signers))
		}
	}
	con, e := ssh.Dial("tcp", fmt.Sprintf("%s:%d", target.address, target.port), config)
	if e != nil {
		return nil, e
	}
	return &ssh.Client{Conn: con}, nil
}

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
