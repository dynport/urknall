package main

import (
	"os/exec"
	"path"
	"strings"
)

type Init struct {
	base
}

func (init *Init) Run() error {
	for _, name := range assetNames() {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if strings.HasPrefix(name, "command") {
			e := init.writeAsset(name)
			if e != nil {
				logger.Printf("ERROR writing asset %s: %s", name, e)
			}
		}
	}
	e := init.writeAsset("main.go")
	if e != nil {
		logger.Printf("ERROR writing asset %s: %s", "main", e)
	}
	return nil
}

func (init *Init) build() error {
	b, e := exec.Command("bash", "-xec", "cd "+init.baseDir()+" && go get . && "+path.Base(init.baseDir())).CombinedOutput()
	if e != nil {
		logger.Print(string(b))
		return e
	}
	logger.Print(string(b))
	return nil
}
