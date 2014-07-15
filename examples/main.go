package main

import (
	"log"
	"os"

	"github.com/dynport/urknall"
)

var logger = log.New(os.Stderr, "", 0)

func main() {
	if e := run(); e != nil {
		logger.Fatal(e)
	}
}

type Template struct {
}

func (tpl *Template) Render(p urknall.Package) {
	p.AddCommands("hello", Shell("echo hello world"))
}

func run() error {
	defer urknall.OpenLogger(os.Stdout).Close()
	var target urknall.Target
	var e error
	uri := "ubuntu@my.host"
	password := ""
	if password != "" {
		target, e = urknall.NewSshTargetWithPassword(uri, password)
	} else {
		target, e = urknall.NewSshTarget(uri)
	}
	if e != nil {
		return e
	}
	return urknall.Run(target, &Template{})
}
