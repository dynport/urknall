package zwo

import (
	"fmt"
	"github.com/dynport/gologger"
	"github.com/dynport/zwo/fw"
)

// The host type. Use the "NewHost" function to create the basic value.
//
// Please note that you need to set the primary interface (the one the host is accessible on) name, if that is not
// "eth0". That should only be necessary on rare circumstances.
//
//	TODO(gfrey): Add better support for interfaces and IPs.
//	TODO(gfrey): Add handling and support for IPv6 (currently the firewall will block everything).
type Host struct {
	IP       string // Host's IP address used to provision the system.
	User     string // User used to log in.
	Hostname string // Hostname used on the system.
	iface    string // Primary network interface of the host.

	Paranoid bool // Make the firewall as restrictive as possible.
	WithVPN  bool // Connect host to a VPN. Assumes "tun0" as interface.

	Docker *DockerSettings // Make the host a docker container carrier.

	Rules  []*fw.Rule  // List of rules used for the firewall.
	IPSets []*fw.IPSet // List of ipsets for the firewall.

	Packages map[string]Package // List of packages of the host.
}

// If the associated host should run (or build) docker containers this type can be used to configure docker.
type DockerSettings struct {
	Version          string // Docker version to run.
	WithRegistry     bool   // Run an image on this host, that will provide a registry for docker images.
	WithBuildSupport bool   // Configure the associated host so that building images is possible.
	Registry         string // URL of the registry to use.
}

// Get the user used to access the host. If none is given the "root" account is as default.
func (h *Host) user() string {
	if h.User == "" {
		return "root"
	}
	return h.User
}

// Set the host's primary interface. Must only be set if the primary interface is not "eth0".
func (h *Host) SetInterface(iface string) {
	h.iface = iface
}

// Get the host's primary interface. If none is given "eth0" is returned as default.
func (h *Host) Interface() string {
	if h.iface == "" {
		return "eth0"
	}
	return h.iface
}

// Add the given package with the given name to the host.
func (h *Host) AddPackage(name string, pkg Package) (e error) {
	if h.Packages == nil {
		h.Packages = map[string]Package{}
	}

	if _, found := h.Packages[name]; found {
		return fmt.Errorf("package with name %q already there", name)
	}

	h.Packages[name] = pkg
	return nil
}

// Provision the host.
func (h *Host) Provision() (e error) {
	return h.provision(false)
}

// Test provisioning of the host, but don't actually execute any commands.
func (h *Host) ProvisionDryrun() (e error) {
	return h.provision(true)
}

func (h *Host) provision(dryrun bool) (e error) {
	packages := make([]Package, 0, len(h.Packages))
	for _, pkg := range h.Packages {
		packages = append(packages, pkg)
	}
	sc := newSSHClient(h)
	if dryrun {
		sc.dryrun = true
		logger.PushPrefix(gologger.Colorize(226, "DRYRUN"))
		defer logger.PopPrefix()
	}
	return sc.provisionHost(packages...)
}

// Provision the given packages into a docker container image tagged with the given tag (the according registry will be
// added automatically). The build will happen on this host, that must be a docker host with build capability.
func (h *Host) CreateDockerImage(tag string, packages ...Package) (imageId string, e error) {
	if !h.IsDockerHost() {
		return "", fmt.Errorf("host %s is not a docker host", h.Hostname)
	}
	dc, e := newDockerClient(h)
	if e != nil {
		return "", e
	}
	return dc.provisionImage(tag, packages...)
}

// Get docker version that should be used. Will panic if the host has no docker enabled.
func (h *Host) DockerVersion() string {
	if h.Docker == nil {
		panic("not a docker host")
	}
	if h.Docker.Version == "" {
		return "0.7.0"
	}
	return h.Docker.Version
}

// Predicate to test whether docker must be installed.
func (h *Host) IsDockerHost() bool {
	return h.Docker != nil
}

// Predicate to test whether the host should be used to build docker images.
func (h *Host) IsDockerBuildHost() bool {
	return h.Docker != nil && h.Docker.WithBuildSupport
}

// Predicate to test whether sudo is required (user for the host is not "root").
func (h *Host) IsSudoRequired() bool {
	if h.User != "" && h.User != "root" {
		return true
	}
	return false
}
