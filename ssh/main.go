package ssh

import (
	"fmt"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	"github.com/dynport/urknall"
)

type Host struct {
	Address  string
	Password string

	address string
	port    int
	user    string

	client *ssh.Client
}

func (host *Host) String() string {
	host.parseAddress()
	return host.address
}

func (host *Host) User() string {
	host.parseAddress()
	return host.user
}

func (host *Host) parseAddress() {
	if host.port > 0 {
		return
	}
	hostAndPort := strings.Split(host.Address, ":")
	var addr string
	if len(hostAndPort) == 2 {
		addr = hostAndPort[0]
	} else {
		host.port = 22
		addr = host.Address
	}
	userAndAddress := strings.Split(addr, "@")
	if len(userAndAddress) == 2 {
		host.user = userAndAddress[0]
		host.address = userAndAddress[1]
	} else {
		host.user = "root"
		host.address = addr
	}

}

type SshClient interface {
	Client() (*ssh.Client, error)
}

func (c *Host) Client() (*ssh.Client, error) {
	c.parseAddress()
	var e error
	config := &ssh.ClientConfig{
		User: c.user,
	}
	if c.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(c.Password))
	}
	con, e := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.address, c.port), config)
	if e != nil {
		return nil, e
	}
	return &ssh.Client{Conn: con}, nil
}

func (c *Host) Command(cmd string) (urknall.Command, error) {
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
