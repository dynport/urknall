package urknall

func newRunlist(name string, pkg Packager, host *Host) *Package {
	return &Package{name: name, pkg: pkg, host: host}
}

type hostPackage struct {
	*Host
	cmds []interface{}
}

func (h *PackageList) newHostPackage(cmds ...interface{}) *hostPackage {
	return &hostPackage{Host: nil, cmds: cmds}
}

func (h *hostPackage) Interface() string {
	return h.publicInterface()
}

func (hp *hostPackage) Package(rl *Package) {
	for i := range hp.cmds {
		rl.Add(hp.cmds[i])
	}
}
