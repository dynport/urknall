package main

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/pkg/nginx"
)

func provisionHost() {
	host := &urknall.Host{
		IP:       "127.0.0.1",
		Hostname: "my-urknall-host",
		User:     "root",
	}

	// run commands, "upgrade" is the name which is used to cache execution
	host.AddCommands("upgrade", "apt-get update", "apt-get upgrade -y")

	// write files
	host.AddCommands("marker", cmd.WriteFile("/tmp/installed.txt", "OK", "root", 0644))

	// install packages (implementing urknall.Package)
	host.AddPackage("nginx", nginx.New("1.4.4"))

    // provision host with ssh and no extra options
	host.Provision(nil)
}
