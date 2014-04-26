package cmd

import "io"

// The Command interface is used to have specialized commands that are used for execution and logging (the latter is
// useful to hide the gory details of more complex commands).
type Command interface {
	Shell() string   // Used for executing the action in a shell (locally or via ssh).
	Logging() string // Get string used for logging.
}

// Interface that allows for rendering template content into a structure. Implement this interface for commands that
// should have the ability for templating. For example the ShellCommand provided by `urknall init` implements this,
// allowing for substitution of a package's values in the command.
type Renderer interface {
	Render(i interface{})
}

// Interface used for types that will validate its state. An error is returned if the state is invalid. Implement this
// on commands to verify validity.
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
