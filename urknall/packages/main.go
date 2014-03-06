package main

import (
	"log"

	"github.com/dynport/urknall"
)

func main() {
	l := urknall.OpenStdoutLogger()
	defer l.Close()
	host := &urknall.Host{IP: "127.0.0.1", User: "root"}
	e := urknall.Provision(host)
	if e != nil {
		log.Fatal(e)
	}
	log.Print("provisioned host")
}
