package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type templatesAdd struct {
	base
	tplBase

	Names []string `cli:"arg required"`
}

func (a *templatesAdd) Run() error {
	tmpls, e := allUpstreamTemplates(a.Repo, a.RepoPath)
	if e != nil {
		return e
	}
	notExisting := []string{}
	for _, name := range a.Names {
		if !tmpls.exists(name) {
			notExisting = append(notExisting, name)
		}
	}
	if len(notExisting) > 0 {
		return fmt.Errorf("template %q does not exist. Existing names %q", notExisting, tmpls.names())
	}

	for _, name := range a.Names {
		e = tmpls[name].Load()
		if e != nil {
			return e
		}
		content, e := tmpls[name].DecodedContent()
		if e != nil {
			return e
		}

		if e = ioutil.WriteFile(filepath.Join(a.BaseDir, "tpl_"+name+".go"), content, 0644); e != nil {
			return e
		}
	}
	return nil
}
