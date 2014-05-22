package main

import (
	"log"
	"os"

	"github.com/dynport/urknall"
)

type Base struct {
}

func (b *Base) Render(p urknall.Package) {
	p.AddCommands("base", Shell("echo hello world"))
}

func main() {
	defer urknall.OpenLogger(os.Stdout).Close()
	target, e := urknall.NewSshTarget("ubuntu@127.0.0.1")
	if e != nil {
		log.Fatal(e)
	}
	pkg := &Base{}
	e = urknall.Run(target, pkg)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("provisioned host")
}
