package urknall

import (
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/local"
	"github.com/dynport/urknall/ssh"
)

type Target interface {
	Command(cmd string) (cmd.ExecCommand, error)
	User() string
	String() string
}

func NewSshTarget(address string) (Target, error) {
	return ssh.New(address)
}

func NewSshTargetWithPassword(address, password string) (Target, error) {
	target, e := ssh.New(address)
	if e == nil {
		target.Password = password
	}
	return target, e
}

func NewLocalTarget() (Target, error) {
	return local.New(), nil
}
