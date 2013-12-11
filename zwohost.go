package zwo

import (
	"fmt"
	"github.com/dynport/zwo/fw"
	"net"
)

// The host type. Use the "NewHost" function to create the basic value.
//
// Please note that you need to set the primary interface (the one the host is accessible on) name, if that is not
// "eth0". That should only be necessary on rare circumstances.
//
//	TODO(gfrey): Add better support for interfaces and IPs.
//	TODO(gfrey): Add handling and support for IPv6 (currently the firewall will block everything).
type Host struct {
	ip       net.IP // Host's IP address used to provision the system.
	user     string // User used to log in.
	hostname string // Hostname used on the system.
	iface    string // Primary network interface of the host.

	Paranoid bool // Make the firewall as restrictive as possible.
	WithVPN  bool // Connect host to a VPN. Assumes "tun0" as interface.

	Docker *DockerSettings // Make the host a docker container carrier.

	Rules  []*fw.Rule  // List of rules used for the firewall.
	IPSets []*fw.IPSet // List of ipsets for the firewall.
}

// If the associated host should run (or build) docker containers this type can be used to configure docker.
type DockerSettings struct {
	Version          string // Docker version to run.
	WithRegistry     bool   // Run an image on this host, that will provide a registry for docker images.
	WithBuildSupport bool   // Configure the associated host so that building images is possible.
	Registry         string // URL of the registry to use.
}

// Create a new hosts structure. This function will return an error if no or and invalid IP (ipv4 and ipv6 are
// supported) is given. The hostname is optional, but should be set for better communication on the host. The user set
// in here should be the user used to access the machine. As SSH is used with publickey authentication, this user must
// exist on the host and have one of your public keys in its authorizes hosts file. If an empty user is given, "root" is
// assumed.
func NewHost(ip, user, hostname string) (host *Host, e error) {
	h := &Host{Paranoid: true}

	if ip == "" {
		return nil, fmt.Errorf("no IP address given")
	}

	e = h.setIPAddress(ip)
	if e != nil {
		return nil, e
	}

	h.user = user
	h.hostname = hostname

	return h, nil
}

// Set the host's IP address.
func (h *Host) setIPAddress(ip string) (e error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("not a valid IP address (must be either IPv4 or IPv6): %s", ip)
	}
	h.ip = parsedIP
	return nil
}

// Get the host's IP address. This is the public IP address used for SSH access.
func (h *Host) IPAddress() string {
	return h.ip.String()
}

// Get the host's name.
func (h *Host) Hostname() string {
	return h.hostname
}

// Get the user used to access the host. If none is given the "root" account is as default.
func (h *Host) User() string {
	if h.user == "" {
		return "root"
	}
	return h.user
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
	if h.user != "" && h.user != "root" {
		return true
	}
	return false
}
