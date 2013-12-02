package zwo

// The action interface must be implemented by the different action types. It is used to run the command in different
// contexts, i.e. running a command in the shell is different from docker and logging wise handling commands is helpful
// too (writing files shouldn't print the underlying shell commands used).
type action interface {
	Docker() string  // Used for executing the action in a docker context.
	Shell() string   // Used for executing the action in a shell (locally or via ssh).
	Logging() string // Get string used for logging.
}
