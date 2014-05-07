package urknall

type Build struct {
	Target
	Pkg    *Package
	DryRun bool
	Env    []string
}

func (build Build) Run() error {
	e := build.Pkg.precompileRunlists()
	if e != nil {
		return e
	}

	e = build.prepare()
	if e != nil {
		return e
	}
	return build.run()
}
