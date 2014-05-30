package urknall

// The Command interface must be implemented by commands. There are different
// commands available in the templates provided by the urknall binary. It's
// also possible to write your own commands, of course.
type Command interface {
	Shell() string   // Used for executing the action in a shell (locally or via ssh).
	Logging() string // Get string used for logging.
}

type renderer interface {
	Render(i interface{})
}
