package urknall

import (
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/target"
)

type Target interface {
	Command(cmd string) (cmd.ExecCommand, error)
	User() string
	String() string
	Reset() error
}

func NewSshTarget(address string) (Target, error) {
	return target.NewSshTarget(address)
}

func NewSshTargetWithPassword(address, password string) (Target, error) {
	target, e := target.NewSshTarget(address)
	if e == nil {
		target.Password = password
	}
	return target, e
}

func NewLocalTarget() (Target, error) {
	return target.NewLocalTarget(), nil
}
