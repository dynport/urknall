package urknall

func Provision(commander Commander, list *PackageList) error {
	return ProvisionWithOptions(commander, list, nil)
}

type userer interface {
	User() string
}

type ProvisionOptions struct {
	DryRun bool
}

func ProvisionWithOptions(commander Commander, list *PackageList, opts *ProvisionOptions) error {
	e := list.precompileRunlists()
	if e != nil {
		return e
	}

	runner := &Runner{
		Commander: commander,
	}

	if userer, ok := commander.(userer); ok {
		runner.User = userer.User()
	} else {
		runner.User = "root"
	}

	e = prepareHost(runner)
	if e != nil {
		return e
	}
	return provisionRunlists(list.runlists(), runner)
}
