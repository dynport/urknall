package zwo

// The action interface must be implemented by the different action types.
type action interface {
	Docker() string  // Used for executing the action in a docker context.
	Shell() string   // Used for executing the action in a shell (locally or via ssh).
	Logging() string // Get string used for logging.
}
