package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type base struct {
	BaseDir string `cli:"opt --base-dir"`
}

func (base *base) baseDir() string {
	p := base.BaseDir
	if p == "" {
		p = "."
	}
	abs, e := filepath.Abs(p)
	if e != nil {
		panic(e.Error())
	}
	return abs
}

func (init *base) writeAsset(name string) error {
	dst := init.baseDir() + "/uk_" + name
	logger.Printf("writing asset %q to %q", name, dst)
	dir := path.Dir(dst)
	_, e := os.Stat(dir)
	if e != nil {
		logger.Printf("creating directory %q", dir)
		e = os.Mkdir(dir, 0755)
		if e != nil {
			return e
		}
	}
	b, e := readAsset(name)
	if e != nil {
		return fmt.Errorf("unable to read asset %s: %s", name, e.Error())
	}
	return ioutil.WriteFile(dst, b, 0644)
}
