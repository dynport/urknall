package main

import (
	"fmt"
	"os"

	"path/filepath"
	"strings"
)

type initProject struct {
	Repo     string `cli:"opt -r --repo default=dynport/urknall desc='repository used to retrieve files from'"`
	RepoPath string `cli:"opt -p --path default=examples desc='path in repository used to retrieve files from'"`

	BaseDir string `cli:"arg required"`
}

func (init *initProject) Run() error {
	dir, e := filepath.Abs(init.BaseDir)
	if e != nil {
		return e
	}

	_, e = os.Stat(dir)
	switch {
	case os.IsNotExist(e):
		if e = os.Mkdir(dir, 0755); e != nil {
			return e
		}
	case e != nil:
		return e
	}

	contents, err := upstreamFiles(init.Repo, init.RepoPath)
	if err != nil {
		return fmt.Errorf("loading upstream files: %s", err)
	}

	for _, c := range contents {
		localPath := dir + "/" + c.Name
		if strings.HasPrefix(c.Name, "cmd_") || c.Name == "main.go" {
			_, e := os.Stat(localPath)
			switch {
			case e == nil:
				logger.Printf("file %q exists", c.Name)
				continue
			case os.IsNotExist(e):
				e = c.Load()
				if e != nil {
					return e
				}

				content, e := c.DecodedContent()
				if e != nil {
					return e
				}
				e = writeFile(localPath, content)
				if e != nil {
					return e
				}

				logger.Printf("created %q", c.Name)
			default:
				return e
			}
		}
	}
	return nil
}
