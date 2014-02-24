package main

// The Command interface must be implemented by the different command types. It is used to run the command in different
// contexts, i.e. either shell, docker, or logging. The last one comes in handy if a command's underlying actions are
// rather lengthy or cryptic, but the intent is described easily (like writing assets or files for example).
type Command interface {
	Shell() string   // Used for executing the action in a shell (locally or via ssh).
	Logging() string // Get string used for logging.
}

type DockerCommand interface {
	Docker() string // Used for executing the action in a docker context.
}

type Renderer interface {
	Render(i interface{})
}

type Validator interface {
	Validate() error
}
