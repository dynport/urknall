package main

import "fmt"

type templatesList struct {
	base
}

func (list *templatesList) Run() error {
	all, e := allUpstreamTemplates(list.Repo, list.RepoPath)
	if e != nil {
		return e
	}
	fmt.Println("available templates: ")
	for _, name := range all.names() {
		fmt.Println(" * " + name)
	}
	return nil
}
