package main

import (
	"log"

	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/runner/ssh"
)

func provisionHost() {
	host := &ssh.Host{
		Address: "127.0.0.1",
	}

	list := &urknall.PackageList{}

	// run commands, "upgrade" is the name which is used to cache execution
	list.AddCommands("upgrade", "apt-get update", "apt-get upgrade -y")

	// write files
	list.AddCommands("marker", cmd.WriteFile("/tmp/installed.txt", "OK", "root", 0644))

	// install packages (implementing urknall.Package)
	list.AddPackage("nginx", &Nginx{Version: "1.4.4"})

	// provision host with ssh and no extra options

	if e := urknall.Provision(host, list); e != nil {
		log.Fatal(e)
	}
}
