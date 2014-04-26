package urknall

import "github.com/dynport/urknall/cmd"

type Host interface {
	Command(cmd string) (cmd.ExecCommand, error)
	User() string
	String() string
}
