package main

import (
	"os"

	"path/filepath"
	"strings"
)

type initProject struct {
	base

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

	contents, e := upstreamFiles(init.Repo, init.RepoPath)
	if e != nil {
		return e
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
