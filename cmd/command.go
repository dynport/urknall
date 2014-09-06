// The Command Interfaces
//
// This package contains a set of interfaces, commands must or can implement.
package cmd

import "io"

// The Command interface is used to have specialized commands that are used for
// execution and logging (the latter is useful to hide the gory details of more
// complex commands).
type Command interface {
	Shell() string
}

// The Logger interface should be implemented by commands, which hide their
// intent behind a series of complex shell commands. The returned string will
// be printed instead of the raw output of the Shell function.
type Logger interface {
	Logging() string
}

// If a command needs to send something to the remote host (a file for example)
// the content can be made available on standard input of the remote command.
// The command must make sure that changed local content will reissue execution
// of the command (by printing the content's hash to standard output for
// example).
type StdinConsumer interface {
	Input() io.ReadCloser
}

// Often it is convenient to directly use values or methods of the template in
// the commands (using go's templating mechanism).
type Renderer interface {
	Render(i interface{})
}

// Interface used for types that will validate its state. An error is returned
// if the state is invalid.
type Validator interface {
	Validate() error
}
