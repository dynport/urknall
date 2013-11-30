package zwo

import (
	"fmt"
	"github.com/dynport/zwo/host"
	"runtime/debug"
)

// Create a new provisioner for the given host.
func ProvisionHost(host *host.Host, packages ...Compiler) (e error) {
	return newSSHClient(host).Provision(packages...)
}

func ProvisionImage(host *host.Host, tag string, packages ...Compiler) (imageId string, e error) {
	if !host.IsDockerHost() {
		return "", fmt.Errorf("host %s is not a docker host", host.Hostname())
	}
	dc, e := newDockerClient(host)
	if e != nil {
		return "", e
	}
	return dc.CreateImage(tag, packages...)
}

// Precompile the given packages for the given host.
func precompileRunlists(host *host.Host, packages ...Compiler) (runLists []*Runlist, e error) {
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

	for _, pkg := range packages { // Precompile runlists.
		pkgName := getPackageName(pkg)

		rl := &Runlist{host: host}
		rl.setConfig(pkg)
		rl.setName(pkgName)
		pkg.Compile(rl)

		runLists = append(runLists, rl)
		logger.Debugf("Precompiled package %s", pkgName)
	}

	return runLists, nil
}

// Provision the given list of runlists.
func provisionRunlists(runLists []*Runlist, provisionFunc func(*Runlist) error) (e error) {
	for i := range runLists {
		rl := runLists[i]

		logger.PushPrefix(padToFixedLength(rl.getName(), 15))

		if e = provisionFunc(runLists[i]); e != nil {
			logger.Errorf("failed to provision: %s", e.Error())
			return e
		}

		logger.PopPrefix()
	}
	return nil
}
