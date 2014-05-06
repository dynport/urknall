package urknall

type Build struct {
	Target
	*Package
	DryRun bool
}

func (p *Build) Run() error {
	e := p.Package.precompileRunlists()
	if e != nil {
		return e
	}

	runner := &Runner{Target: p.Target}
	e = runner.prepare()
	if e != nil {
		return e
	}
	return runner.provision(p.Package)
}
