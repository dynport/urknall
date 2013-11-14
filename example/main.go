// An example demonstrating usage of zwo.
//
// The ExamplePackage struct is use to configure a custom package. This package is provisioned together with the
// BasePackage on a target host with the public IP '134.119.1.181'.
package main

import (
	"github.com/dynport/zwo/host"
	"github.com/dynport/zwo/pkg/base"
	"github.com/dynport/zwo/zwo"
)

type ExamplePackage struct {
	Version string `json:"version" default:"0.0.1" required:"true"`
}

func (ex *ExamplePackage) Compile(r *zwo.Runlist) (e error) {
	e = r.AddCommands(zwo.Execute("echo {{ .Version }} >> /tmp/version"))
	return e
}

func main() {
	pkgs := []zwo.Compiler{
		&ExamplePackage{
			Version: "0.0.2",
		},
		&base.BasePackage{
			Packages:    []string{"vim"},
			TimezoneUTC: true,
			SwapSize:    "1000",
			Limits:      true,
		},
	}

	h, e := host.NewHost(host.HOST_TYPE_SSH)
	if e != nil {
		panic(e)
	}
	h.SetPublicIPAddress("134.119.1.181")
	h.SetUser("gfrey")

	provisioner := zwo.NewProvisioner(h)
	if e := provisioner.Provision(pkgs...); e != nil {
		panic(e)
	}
}
