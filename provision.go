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

func Provision(host Target, list *Package) error {
	return ProvisionWithOptions(host, list, nil)
}

type ProvisionOptions struct {
	DryRun bool
}

func ProvisionWithOptions(host Target, list *Package, opts *ProvisionOptions) error {
	e := list.precompileRunlists()
	if e != nil {
		return e
	}

	runner := &Runner{
		Target: host,
	}
	e = runner.prepare()
	if e != nil {
		return e
	}
	return runner.provision(list)
}
