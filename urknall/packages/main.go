package main

import (
	"log"

	"github.com/dynport/urknall"
	"github.com/dynport/urknall/ssh"
)

func main() {
	l := urknall.OpenStdoutLogger()
	defer l.Close()
	host := &ssh.Host{Address: "127.0.0.1:22"}
	e := urknall.Provision(host, &urknall.PackageList{})
	if e != nil {
		log.Fatal(e)
	}
	log.Print("provisioned host")
}
