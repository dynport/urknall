package zwo

import (
	"fmt"
	"runtime/debug"
)

// Precompile the given packages for the given host.
func precompileRunlists(host *Host, packages ...Package) (runLists []*Runlist, e error) {
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
