package urknall

import "io"

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
