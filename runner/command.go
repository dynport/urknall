package runner

import "io"

type Command interface {
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	SetStdin(io.Reader)
	Run() error
}
