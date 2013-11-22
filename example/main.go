// An example demonstrating usage of zwo.
//
// The ExamplePackage struct is use to configure a custom package. This package is provisioned together with the
// BasePackage on a target host with the public IP '134.119.1.181'.
package main

import (
	"fmt"
	"github.com/dynport/gocli"
	. "github.com/dynport/zwo/cmd"
	"github.com/dynport/zwo/host"
	"github.com/dynport/zwo/pkg/base"
	"github.com/dynport/zwo/pkg/docker"
	"github.com/dynport/zwo/zwo"
	"os"
)

type ExamplePackage struct {
	Version string `json:"version" default:"0.0.1" required:"true"`
}

func (ex *ExamplePackage) Compile(r *zwo.Runlist) {
	r.Execute(And(UpdatePackages(), InstallPackages("netcat.traditional")))
	r.Init(nil, "/bin/nc -l -p 2000 -c 'xargs -n1 echo'")
}

type DockerRunner struct {
	ImageId string
}

func (d *DockerRunner) Compile(r *zwo.Runlist) {
	r.Execute("docker run -d -p 0.0.0.0:2000:2000 " + d.ImageId)
}

func provisionHost(args *gocli.Args) (e error) {
	pkgs := []zwo.Compiler{
		&base.BasePackage{
			TimezoneUTC: true,
		},
		&base.Firewall{
			PrimaryInterface: "eth0",
			WithDHCP:         true,
			Paranoid:         true,
			Services: []*base.FWService{
				{
					Description: "Inbound access to Docker API",
					Chain:       "INPUT",
					Port:        4243,
					Interface:   "eth0",
					Protocols:   "tcp",
				},
				{
					Description: "Inbound access to Docker Registry",
					Chain:       "INPUT",
					Port:        5000,
					Interface:   "eth0",
					Protocols:   "tcp",
				},
			},
		},
		&docker.Host{
			Version:      "0.6.7",
			Public:       true,
			Debug:        true,
			WithRegistry: true,
		},
	}

	h, e := host.NewHost(host.HOST_TYPE_SSH)
	if e != nil {
		panic(e)
	}
	h.SetUser("gfrey")
	h.SetPublicIPAddress("192.168.1.20")

	provisioner := zwo.NewProvisioner(h)
	return provisioner.Provision(pkgs...)
}

func provisionDockerContainer(args *gocli.Args) (e error) {
	h, e := host.NewHost(host.HOST_TYPE_DOCKER)
	if e != nil {
		panic(e)
	}
	h.SetPublicIPAddress("192.168.1.20")

	provisioner := zwo.NewProvisioner(h)
	return provisioner.Provision(&ExamplePackage{Version: "0.0.1"})
}

func runDockerImage(args *gocli.Args) (e error) {
	imageId := args.Args[0]

	h, e := host.NewHost(host.HOST_TYPE_SSH)
	if e != nil {
		panic(e)
	}
	h.SetUser("gfrey")
	h.SetPublicIPAddress("192.168.1.20")

	provisioner := zwo.NewProvisioner(h)
	return provisioner.Provision(&DockerRunner{ImageId: imageId})
}

func main() {
	r := gocli.NewRouter(nil)
	r.Register("host/provision", &gocli.Action{Handler: provisionHost})
	r.Register("container/create", &gocli.Action{Handler: provisionDockerContainer})
	r.Register("container/run", &gocli.Action{Handler: runDockerImage, Usage: "image"})

	if e := r.Handle(os.Args); e != nil {
		fmt.Println(e.Error())
		os.Exit(1)
	}
}
