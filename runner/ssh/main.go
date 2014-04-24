package ssh

import (
	"log"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	"github.com/dynport/urknall"
)

var debugger = log.New(os.Stderr, "", 0)

type Host struct {
	Address  string
	Password string

	address string
	port    int
	user    string

	client *ssh.Client
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

func (host *Host) User() string {
	host.parseAddress()
	parts := strings.Split(host.Address, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return "root"
}

type SshClient interface {
	Client() (*ssh.Client, error)
}

func (c *Host) Client() (*ssh.Client, error) {
	var e error
	config := &ssh.ClientConfig{
		User: c.User(),
	}
	if c.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(c.Password))
	}
	addr := c.Address
	if !strings.Contains(addr, ":") {
		addr += ":22"
	}
	debugger.Printf("connecting %q with %#v", addr, config)
	con, e := ssh.Dial("tcp", addr, config)
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
