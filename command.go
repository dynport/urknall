package urknall

import (
	"io"

	"github.com/dynport/urknall/utils"
)

type stringCommand struct {
	cmd string
}

func (sc *stringCommand) Shell() string {
	return sc.cmd
}

func (sc *stringCommand) Logging() string {
	return "[COMMAND] " + sc.cmd
}

func (sc *stringCommand) Render(i interface{}) {
	sc.cmd = utils.MustRenderTemplate(sc.cmd, i)
}

type Command interface {
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	StdinPipe() (io.Writer, error)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	SetStdin(io.Reader)
	Run() error
	Start() error
	Wait() error
}
