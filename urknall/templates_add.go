package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type templatesAdd struct {
	*base
	Names []string `cli:"arg required"`
}

func (a *templatesAdd) Run() error {
	var e error
	a.base, e = loadBase()
	if e != nil {
		return e
	}
	tmpls, e := allTemplates()
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

		if e = ioutil.WriteFile("tpl_"+name+".go", content, 0644); e != nil {
			return e
		}
	}
	return nil
}

func allTemplates() (tmpls templates, e error) {
	tmpls = templates{}
	contents, e := exampleFiles()
	if e != nil {
		return nil, e
	}

	for _, c := range contents {
		if strings.HasPrefix(c.Name, "tpl_") && strings.HasSuffix(c.Name, ".go") {
			name := c.Name[4 : len(c.Name)-3]
			tmpls[name] = c
		}
	}
	return tmpls, nil
}
