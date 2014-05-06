package ssh

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"github.com/dynport/urknall/cmd"
)

func New(addr string) (host *Host, e error) {
	host = &Host{port: 22, user: "root"}

	hostAndPort := strings.SplitN(addr, ":", 2)
	if len(hostAndPort) == 2 {
		addr = hostAndPort[0]
		host.port, e = strconv.Atoi(hostAndPort[1])
		if e != nil {
			return nil, fmt.Errorf("port must be given as integer, got %q", hostAndPort[1])
		}
	}

	userAndAddress := strings.Split(addr, "@")
	switch len(userAndAddress) {
	case 1:
		host.address = addr
	case 2:
		host.user = userAndAddress[0]
		host.address = userAndAddress[1]
	default:
		return nil, fmt.Errorf("expected host address of the form '<user>@<host>', but was given: %s", addr)
	}

	if host.address == "" {
		e = fmt.Errorf("empty address given for host")
	}

	return host, e
}

type Host struct {
	Password string

	user    string
	port    int
	address string

	client *ssh.Client
}

func (host *Host) User() string {
	return host.user
}

func (host *Host) String() string {
	return fmt.Sprintf("%s@%s:%d", host.user, host.address, host.port)
}

type SshClient interface {
	Client() (*ssh.Client, error)
}

func (c *Host) Client() (*ssh.Client, error) {
	var e error
	config := &ssh.ClientConfig{
		User: c.user,
	}
	if c.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(c.Password))
	} else if sshSocket := os.Getenv("SSH_AUTH_SOCK"); sshSocket != "" {
		if c, e := net.Dial("unix", sshSocket); e == nil {
			config.Auth = append(config.Auth, ssh.PublicKeysCallback(agent.NewClient(c).Signers))
		}
	}
	con, e := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.address, c.port), config)
	if e != nil {
		return nil, e
	}
	return &ssh.Client{Conn: con}, nil
}

func (c *Host) Command(cmd string) (cmd.ExecCommand, error) {
	if c.client == nil {
		var e error
		c.client, e = c.Client()
		if e != nil {
			return nil, e
		}
	}
	ses, e := c.client.NewSession()
	if e != nil {
		return nil, e
	}
	return &Command{command: cmd, session: ses}, nil
}
