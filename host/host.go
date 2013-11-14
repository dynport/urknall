// struct to provide basic information on the host to provision.
//
// Depending on the type of host different mechanisms and information are required. This knowledge is encapsulated in
// the Host struct.
package host

import (
	"fmt"
	"net"
)

type HostType int

const (
	HOST_TYPE_DOCKER HostType = iota // Host is a docker image.
	HOST_TYPE_SSH    HostType = iota // Host is a machine accessible using SSH.
)

type Host struct {
	hostType int    // What executor should be used (SSH or Docker)?
	publicIP net.IP // Host's IP address used to provision the system.
	vpnIP    net.IP // Host's private IP address.
	user     string // User used to log in.
}

// Create a new host of the given type.
func NewHost(hostType HostType) (host *Host, e error) {
	if hostType != HOST_TYPE_SSH && hostType != HOST_TYPE_DOCKER {
		return nil, fmt.Errorf("host type must be of the HOST_TYPE_{DOCKER,SSH} const")
	}
	return &Host{hostType: hostType}, nil
}

// Returns true if this host is a docker image.
func (h *Host) IsDockerHost() bool {
	return h.hostType == HOST_TYPE_DOCKER
}

// Returns true if this host is accessible using SSH.
func (h *Host) IsSshHost() bool {
	return h.hostType == HOST_TYPE_SSH
}

// Set the public IP of the host.
//
// TODO This only makes sense for SSH hosts.
func (h *Host) SetPublicIPAddress(ip string) (e error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("not a valid IP address (either IPv4 or IPv6): %s", ip)
	}
	h.publicIP = parsedIP
	return nil
}

// Get the public IP address of the host.
func (h *Host) GetPublicIPAddress() string {
	if h.publicIP == nil {
		return ""
	}
	return h.publicIP.String()
}

// Set the IP of the host inside a VPN.
//
// TODO This only makes sense for SSH hosts.
func (h *Host) SetVpnIPAddress(ip string) (e error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("not a valid IP address (either IPv4 or IPv6): %s", ip)
	}
	h.vpnIP = parsedIP
	return nil
}

// Get the VPN IP address of the host.
func (h *Host) GetVpnIPAddress() string {
	if h.vpnIP == nil {
		return ""
	}
	return h.vpnIP.String()
}

// Set the user used to access the host. If none is given the 'root' account is as default.
//
// TODO This only makes sense for SSH hosts.
func (h *Host) SetUser(user string) {
	h.user = user
}

// Get the user used to access the host. If none is given the 'root' account is as default.
func (h *Host) GetUser() string {
	if h.user == "" {
		return "root"
	}
	return h.user
}

// Predicate to test whether sudo is required (user for the host is not 'root').
//
// TODO This only makes sense for SSH hosts.
func (h *Host) IsSudoRequired() bool {
	if h.user != "" && h.user != "root" {
		return true
	}
	return false
}
