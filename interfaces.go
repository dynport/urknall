package urknall

type Command interface {
	Shell() string   // Used for executing the action in a shell (locally or via ssh).
	Logging() string // Get string used for logging.
}

type Renderer interface {
	Render(i interface{})
}
