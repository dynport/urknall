// struct to provide basic information on the host to provision.
package host

import (
	"fmt"
	"net"
)

type Host struct {
	ip       net.IP // Host's IP address used to provision the system.
	user     string // User used to log in.
	hostname string // Hostname used on the system.
	iface    string // Primary network interface of the host.

	Paranoid bool // Make the firewall as restrictive as possible.
	WithVPN  bool // Connect host to a VPN. Assumes 'tun0' as interface.
}

// Create a new hosts structure.
func New(ip, user, hostname string) (host *Host, e error) {
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
		return fmt.Errorf("not a valid IP address (either IPv4 or IPv6): %s", ip)
	}
	h.ip = parsedIP
	return nil
}

// Get the host's IP address.
func (h *Host) IPAddress() string {
	return h.ip.String()
}

// Get the host's name.
func (h *Host) Hostname() string {
	return h.hostname
}

// Get the user used to access the host. If none is given the 'root' account is as default.
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

// Predicate to test whether sudo is required (user for the host is not 'root').
func (h *Host) IsSudoRequired() bool {
	if h.user != "" && h.user != "root" {
		return true
	}
	return false
}
