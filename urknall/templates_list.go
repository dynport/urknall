package main

import "fmt"

type templatesList struct {
	Repo     string `cli:"opt -r --repo default=dynport/urknall desc='repository used to retrieve files from'"`
	RepoPath string `cli:"opt -p --path default=examples desc='path in repository used to retrieve files from'"`
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
