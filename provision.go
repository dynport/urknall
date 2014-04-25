package urknall

func Provision(host Host, list *PackageList) error {
	return ProvisionWithOptions(host, list, nil)
}

type userer interface {
	User() string
}

type ProvisionOptions struct {
	DryRun bool
}

func ProvisionWithOptions(host Host, list *PackageList, opts *ProvisionOptions) error {
	e := list.precompileRunlists()
	if e != nil {
		return e
	}

	runner := &Runner{
		Host: host,
	}

	if userer, ok := host.(userer); ok {
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
