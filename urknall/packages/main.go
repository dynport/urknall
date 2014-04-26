package main

import (
	"log"

	"github.com/dynport/urknall"
	"github.com/dynport/urknall/ssh"
)

func main() {
	defer urknall.OpenStdoutLogger().Close()
	list := &urknall.PackageList{}
	list.Add("pkg.hello", "echo hello world")
	host := &ssh.Host{Address: "ubuntu@127.0.0.1"}
	e := urknall.Provision(host, list)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("provisioned host")
}
