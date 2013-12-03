package firewall

import (
	"fmt"
	"net"
	"strconv"
)

type hostSet struct {
	Name         string
	WithProtocol bool
	WithPort     bool
}

type Rule struct {
	description string
	chain       string

	protocol string

	srcIface string
	srcPort  int
	srcIP    net.IP
	srcHosts *hostSet

	destIface string
	destPort  int
	destIP    net.IP
	destHosts *hostSet
}

// Add a local service, i.e. something that runs on the local machine an should be made accessible to the outside world.
func LocalService(description string) (r *Rule) {
	return &Rule{description: description, chain: "INPUT"}
}

// Add a remote service, i.e. something that runs outside the box but should be reachable (required for paranoid mode).
func RemoteService(description string) (r *Rule) {
	return &Rule{description: description, chain: "OUTPUT"}
}

// A service provided by docker.
func DockerService(description string) (r *Rule) {
	return &Rule{description: description, chain: "FORWARD", destIface: "docker0"}
}

// A service used by docker.
func ServiceForDocker(description string) (r *Rule) {
	return &Rule{description: description, chain: "FORWARD", srcIface: "docker0"}
}

// For the given rule use the given protocol.
func (r *Rule) WithProtocol(proto string) *Rule {
	r.protocol = proto
	return r
}

// The destination's port.
func (r *Rule) AtDestinationPort(port int) *Rule {
	if r.protocol == "" {
		panic("no protocol selected")
	}
	r.destPort = port
	return r
}

// The source's port.
func (r *Rule) AtSourcePort(port int) *Rule {
	if r.protocol == "" {
		panic("no protocol selected")
	}
	r.srcPort = port
	return r
}

// Service is provided by the given host(s), aka who provides the service. The host type can either be 'ip' or 'set'.
func (r *Rule) ProvidedBy(hostType, host string) *Rule {
	switch hostType {
	case "ip":
		r.destIP = net.ParseIP(host)
	case "set":
		r.destHosts = &hostSet{Name: host}
	default:
		panic(fmt.Sprintf("unknown host type '%s'", hostType))
	}
	return r
}

// Service is provided for the given host(s), aka who can access the service. The host type can either be 'ip' or 'set'.
func (r *Rule) ProvidedFor(hostType, host string) *Rule {
	switch hostType {
	case "ip":
		r.srcIP = net.ParseIP(host)
	case "set":
		r.srcHosts = &hostSet{Name: host}
	default:
		panic(fmt.Sprintf("unknown host type '%s'", hostType))
	}
	return r
}

func (r *Rule) Convert() (cmd string) {
	cmd = "-A " + r.chain
	if r.protocol != "" {
		cmd += " --protocol " + r.protocol
		if r.srcPort != 0 {
			cmd += " -m " + r.protocol + " --source-port " + strconv.Itoa(r.srcPort)
		}
		if r.destPort != 0 {
			cmd += " -m " + r.protocol + " --destination-port " + strconv.Itoa(r.destPort)
		}
	}

	if r.srcHosts != nil {
		cmd += " -m set --match-set " + r.srcHosts.Name + " src"
	}
	if r.destHosts != nil {
		cmd += " -m set --match-set " + r.destHosts.Name + " dst"
	}

	if r.srcIP != nil {
		cmd += " --source " + r.srcIP.String()
	}
	if r.destHosts != nil {
		cmd += " --destination " + r.destIP.String()
	}

	if r.srcIface != "" {
		cmd += " --in-interface " + r.srcIface
	}
	if r.destIface != "" {
		cmd += " --out-interface " + r.destIface
	}

	cmd += " -m state --state NEW"

	cmd += " -j ACCEPT\n"

	return cmd
}
