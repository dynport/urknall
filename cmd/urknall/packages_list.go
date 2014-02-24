package main

import (
	"fmt"
	"sort"
)

type packages map[string]struct{}

func (p packages) exists(name string) bool {
	_, exist := p[name]
	return exist
}

func (p packages) names() []string {
	names := []string{}
	for n := range p {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

type PackagesList struct {
}

func (list *PackagesList) Run() error {
	all := allPackages()
	fmt.Println("available packages: ")
	for _, name := range all.names() {
		fmt.Println("* " + name)
	}
	return nil
}
