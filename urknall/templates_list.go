package main

import (
	"fmt"
	"sort"
)

type templates map[string]*content

func (t templates) exists(name string) bool {
	_, exist := t[name]
	return exist
}

func (t templates) names() []string {
	names := []string{}
	for n := range t {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

type templatesList struct {
}

func (list *templatesList) Run() error {
	all, e := allTemplates()
	if e != nil {
		return e
	}
	fmt.Println("available packages: ")
	for _, name := range all.names() {
		fmt.Println("* " + name)
	}
	return nil
}
