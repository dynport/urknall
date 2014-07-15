package main

import "fmt"

type AddPackage struct {
	*base
	Names []string `cli:"arg required"`
}

func (a *AddPackage) Run() error {
	var e error
	a.base, e = loadBase()
	if e != nil {
		return e
	}
	all := allPackages()
	notExisting := []string{}
	for _, name := range a.Names {
		if !all.exists(name) {
			notExisting = append(notExisting, name)
		}
	}
	if len(notExisting) > 0 {
		return fmt.Errorf("packages %q does not exist. Existing names %q", notExisting, all.names())
	}
	for _, name := range a.Names {
		e := a.writeAsset("pkg_" + name + ".go")
		if e != nil {
			return e
		}
	}
	return nil
}

func allPackages() packages {
	return packages{}
}
