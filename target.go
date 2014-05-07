package urknall

import "github.com/dynport/urknall/cmd"

type Target interface {
	Command(cmd string) (cmd.ExecCommand, error)
	User() string
	String() string
}
