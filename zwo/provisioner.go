package zwo

import (
	"fmt"
	"github.com/dynport/gologger"
	"github.com/dynport/zwo/host"
	"runtime/debug"
)

// Provision the given host with the given packages. If "dryrun" is set, no actions are performed. This can be used to
// get a feeling of what would happen.
func ProvisionHost(host *host.Host, dryrun bool, packages ...Packager) (e error) {
	sc := newSSHClient(host)
	if dryrun {
		sc.dryrun = true
		logger.PushPrefix(gologger.Colorize(226, "DRYRUN"))
		defer logger.PopPrefix()
	}

	return sc.provisionHost(packages...)
}

// Provision the given packages into a docker container image tagged with the given tag (the according registry will be
// added automatically). The build will happen on the given host, that must be a docker host with build capability.
func ProvisionImage(host *host.Host, tag string, packages ...Packager) (imageId string, e error) {
	if !host.IsDockerHost() {
		return "", fmt.Errorf("host %s is not a docker host", host.Hostname())
	}
	dc, e := newDockerClient(host)
	if e != nil {
		return "", e
	}
	return dc.provisionImage(tag, packages...)
}

// Precompile the given packages for the given host.
func precompileRunlists(host *host.Host, packages ...Packager) (runLists []*Runlist, e error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("failed to precompile package: %v", r)
			}
			logger.Info(e.Error())
			logger.Debug(string(debug.Stack()))
		}
	}()

	runLists = make([]*Runlist, 0, len(packages))

	runlistNames := map[string]bool{}

	for _, pkg := range packages { // Precompile runlists.
		if e = validatePackage(pkg); e != nil {
			return nil, e
		}

		pkgName := packageName(pkg)

		if pkgName == "" {
			return nil, fmt.Errorf("package name empty")
		}

		if runlistNames[pkgName] {
			return nil, fmt.Errorf("package %q used twice", pkgName)
		}
		runlistNames[pkgName] = true

		rl := &Runlist{}
		rl.pkg = pkg
		rl.name = pkgName
		pkg.Package(rl)

		runLists = append(runLists, rl)
		logger.Debugf("Precompiled package %s", pkgName)
	}

	return runLists, nil
}

// Provision the given list of runlists.
func provisionRunlists(runLists []*Runlist, provisionFunc func(*Runlist) error) (e error) {
	for i := range runLists {
		rl := runLists[i]

		logger.PushPrefix(padToFixedLength(rl.name, 15))

		if e = provisionFunc(runLists[i]); e != nil {
			logger.Errorf("failed to provision: %s", e.Error())
			return e
		}

		logger.PopPrefix()
	}
	return nil
}
