package urknall

func Build(tgt Target, pkg *Package) error {
	e := pkg.precompileRunlists()
	if e != nil {
		return e
	}

	runner := &Runner{Target: tgt}
	e = runner.prepare()
	if e != nil {
		return e
	}
	return runner.provision(pkg)
}
