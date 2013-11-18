package zwo

import (
	"fmt"
	"github.com/dynport/gossh"
	"github.com/dynport/zwo/host"
	"strings"
)

// Provisioners are responsible for compiling the given packages into runlists and execute those.
type Provisioner interface {
	Provision(packages ...Compiler) (e error)
}

// Create a new provisioner for the given host.
func NewProvisioner(h *host.Host) (p Provisioner) {
	switch {
	case h.IsSshHost():
		sc := gossh.New(h.GetPublicIPAddress(), h.GetUser())
		return &sshClient{client: sc, host: h}
	case h.IsDockerHost():
		return nil
	}
	return nil
}

func getPackageName(pkg Compiler) (name string) {
	pkgName := fmt.Sprintf("%T", pkg)
	return strings.ToLower(pkgName[1:])
}
