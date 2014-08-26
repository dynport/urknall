// Commands are the abstraction of the actual commands being executed on the
// target. Such an abstraction is helpful as they provide the possibility to
// add shortcuts for common tasks and more complex commands. There are some
// interfaces that must or can be implemented, depending on the required
// features.
//
// For further information see http://urknall.dynport.de/docs/library/#commands.
package cmd

import "io"

// All commands must implement this interface. The Shell method returns the
// command actually executed on the target.
type Command interface {
	Shell() string
}

// If not implemented by a command the string returned by the Shell method
// will be used for logging. If this method is implemented the returned string
// will be used instead.
type Logger interface {
	Logging() string
}

type StdinConsumer interface {
	Input() io.ReadCloser
}

// Interface that allows for rendering template content into a structure. Implement this interface for commands that
// should have the ability for templating. For example the ShellCommand provided by `urknall init` implements this,
// allowing for substitution of a package's values in the command.
type Renderer interface {
	Render(i interface{})
}

// This interface can be implemented by commands that need to make sure the
// configuration is valid. This helps to let the command fail as early and
// graceful as possible.
type Validator interface {
	Validate() error
}

type ExecCommand interface {
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
