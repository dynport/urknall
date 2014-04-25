package main

import (
	"fmt"
	"strings"
)

type AddPackage struct {
	base
	Names []string `cli:"arg required"`
}

func (a *AddPackage) Run() error {
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
	m := packages{}
	for _, file := range assetNames() {
		if strings.HasPrefix(file, "pkg_") {
			name := strings.TrimSuffix(strings.TrimPrefix(file, "pkg_"), ".go")
			m[name] = struct{}{}
		}
	}
	return m
}
