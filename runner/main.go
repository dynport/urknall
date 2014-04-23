package runner

type Runner struct {
	User      string
	DryRun    bool
	Env       []string
	Commander Commander
}

type Commander interface {
	Command(cmd string) (Command, error)
}
