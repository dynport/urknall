// struct to provide basic information on the host to provision.
package host

import (
	"fmt"
	"github.com/dynport/zwo/firewall"
	"net"
)

type Host struct {
	ip       net.IP // Host's IP address used to provision the system.
	user     string // User used to log in.
	hostname string // Hostname used on the system.
	iface    string // Primary network interface of the host.

	Paranoid bool // Make the firewall as restrictive as possible.
	WithVPN  bool // Connect host to a VPN. Assumes 'tun0' as interface.

	Docker *DockerSettings // Make the host a docker container carrier.

	rules []*firewall.Rule // List of rules used for the firewall.
}

type DockerSettings struct {
	Version          string
	WithRegistry     bool
	WithBuildSupport bool
	Paranoid         bool
	Registry         string
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
		return fmt.Errorf("not a valid IP address (must be either IPv4 or IPv6): %s", ip)
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

// Get docker version.
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

// Predicate to test whether host should be used to build docker images.
func (h *Host) IsDockerBuildHost() bool {
	return h.Docker != nil && h.Docker.WithBuildSupport
}

// Predicate to test whether sudo is required (user for the host is not 'root').
func (h *Host) IsSudoRequired() bool {
	if h.user != "" && h.user != "root" {
		return true
	}
	return false
}

// Add a firewall rule.
func (h *Host) AddFirewallRule(r *firewall.Rule) {
	h.rules = append(h.rules, r)
}

// Compile rules into something iptables can digest.
func (h *Host) FirewallRules() (rules []string) {
	rules = []string{}
	for _, rule := range h.rules {
		rules = append(rules, rule.Convert())
	}
	return rules
}
