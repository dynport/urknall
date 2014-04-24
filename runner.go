package urknall

type Runner struct {
	User      string
	DryRun    bool
	Env       []string
	Commander Commander
}

func (runner *Runner) IsSudoRequired() bool {
	return runner.User != "root"
}
