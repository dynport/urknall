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

func New(addr string) (target *Target, e error) {
	target = &Target{port: 22, user: "root"}

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

type Target struct {
	Password string

	user    string
	port    int
	address string

	client *ssh.Client
}

func (target *Target) User() string {
	return target.user
}

func (target *Target) String() string {
	return fmt.Sprintf("%s@%s:%d", target.user, target.address, target.port)
}

func (target *Target) Command(cmd string) (cmd.ExecCommand, error) {
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

func (target *Target) buildClient() (*ssh.Client, error) {
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
