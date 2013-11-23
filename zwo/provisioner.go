package zwo

import (
	"fmt"
	"github.com/dynport/dpgtk/docker_client"
	"github.com/dynport/gossh"
	"github.com/dynport/zwo/host"
	"strings"
)

// Provisioners are responsible for compiling the given packages into runlists and execute those.
type Provisioner interface {
	Provision(packages ...Compiler) (e error)
}

// Create a new provisioner for the given host.
func NewSSHProvisioner(h *host.Host) (p Provisioner) {
	sc := gossh.New(h.GetIPAddress(), h.GetUser())
	return &sshClient{client: sc, host: h}
}

func NewDockerProvisioner(h *host.Host) (p Provisioner) {
	c := docker_client.New(h.GetIPAddress())
	return &dockerClient{host: h, client: c}
}

func getPackageName(pkg Compiler) (name string) {
	pkgName := fmt.Sprintf("%T", pkg)
	return strings.ToLower(pkgName[1:])
}
