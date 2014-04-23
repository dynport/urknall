package urknall

func Provision(host *Host, list *PackageList) error {
	return ProvisionWithOptions(host, list, nil)
}

func ProvisionWithOptions(host *Host, list *PackageList, opts *ProvisionOptions) error {
	list.Hostname = host.Hostname
	list.Timezone = host.Timezone

	client := newSSHClient(host, opts)
	e := list.precompileRunlists()
	if e != nil {
		return e
	}

	e = client.prepareHost()
	if e != nil {
		return e
	}
	return provisionRunlists(list.runlists(), client)
}
