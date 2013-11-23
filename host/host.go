// struct to provide basic information on the host to provision.
//
// Depending on the type of host different mechanisms and information are required. This knowledge is encapsulated in
// the Host struct.
package host

import (
	"fmt"
	"net"
)

type Host struct {
	ip   net.IP // Host's IP address used to provision the system.
	user string // User used to log in.
}

func New(ip string) (host *Host, e error) {
	if ip == "" {
		return nil, fmt.Errorf("no IP address given")
	}

	h := &Host{}
	e = h.setIPAddress(ip)
	if e != nil {
		return nil, e
	}

	return h, nil
}

// Set the host's IP address.
func (h *Host) setIPAddress(ip string) (e error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("not a valid IP address (either IPv4 or IPv6): %s", ip)
	}
	h.ip = parsedIP
	return nil
}

// Get the host's IP address.
func (h *Host) GetIPAddress() string {
	return h.ip.String()
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
