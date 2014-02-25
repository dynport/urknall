package main

import (
	"net"

	"github.com/dynport/urknall"
)

type Firewall struct {
	Interface string `urknall:"default=eth0"`
	WithVPN   bool
	Paranoid  bool
	Rules     []*FirewallRule
	IPSets    []*FirewallIPSet // List of ipsets for the firewall.
}

func (f *Firewall) Package(r *urknall.Runlist) {
	r.Add(
		InstallPackages("iptables", "ipset"),
		WriteFile("/etc/network/if-pre-up.d/iptables", firewallUpstart, "root", 0744),
		WriteFile("/etc/iptables/ipsets", fwIpset, "root", 0644),
		WriteFile("/etc/iptables/rules_ipv4", fw_rules_ipv4, "root", 0644),
		WriteFile("/etc/iptables/rules_ipv6", fw_rules_ipv6, "root", 0644),
		"{ modprobe iptable_filter && modprobe iptable_nat; }; /bin/true", // here to make sure next command succeeds.
		"IFACE={{ .Interface }} /etc/network/if-pre-up.d/iptables",
	)
}

// IPSets are the possibility to change a rule, without actually rewriting the rules. That is they add some sort of
// flexibility with regard to dynamic entities like a load balancer, which must have access to the different machines
// that should take the load.
//
// A set is defined by a name, that is used in iptables rule (see "Rule.(Source|Destination).IPSet") to reference the
// contained entities. The type defines what parameters must match an entry (see "ipset --help" output and the man page
// for a list of allowed values), for example a set could define hosts and ports.
//
// The family defines the type of IP address to handle, either IPv4 or IPv6. The allowed values are "inet" and "inet6"
// respectively.
//
// There are some ipset internal parameters that shouldn't need to be changed often. Those are "HashSize" that defines
// the size of the underlying hash. This value defaults to 1024. The "MaxElem" number determines how much elements there
// can be in the set at most.
//
// An initial set of members can be defined, if reasonable.
type FirewallIPSet struct {
	Name     string   // Name of the ipset.
	Type     string   // Type of the ipset.
	Family   string   // Network address family.
	HashSize int      // Size of the hash.
	MaxElem  int      // Max number of elements of the set.
	Members  []net.IP // Initial set of members.
}

// A rule defines what is allowed to flow from some source to some destination. A description can be added to make the
// resulting scripts more readable.
//
// The "Chain" field determines which chain the rule is added to. This should be either "INPUT", "OUTPUT", or "FORWARD",
// with the names of the chains mostly speaking for themselves.
//
// The protocol setting is another easy match for the rule and especially required for some of the target's settings,
// i.e. if a port is specified the protocol must be given too.
//
// Source and destination are the two communicating entities. For the input chain the local host is destination and for
// output it is the source.
type FirewallRule struct {
	Description string
	Chain       string // Chain to add the rule to.
	Protocol    string // The protocol used.

	Source      *FirewallTarget // The source of the packet.
	Destination *FirewallTarget // The destination of the packet.
}

// The target of a rule. It can be specified either by IP or the name of an IPSet. Additional parameters are the port
// and interface used. It's totally valid to only specify a subset (or even none) of the fields. For example IP and
// IPSet must not be given for the host the rule is applied on.
//
// TODO(gfrey): There currently is no validation the referenced IPSet exists. This should be added on provisioning to
// make sure iptables setup won't fail.
type FirewallTarget struct {
	IP        net.IP // IP of the target.
	IPSet     string // IPSet used for matching.
	Port      int    // Port packets must use to match.
	Interface string // Interface the packet goes through.
	NAT       string // NAT configuration (empty, "MASQ", or Interface's IP).
}

const fw_rules_ipv6 = `
*filter
:INPUT DROP [0:0]
:FORWARD DROP [0:0]
:OUTPUT DROP [0:0]

COMMIT
`

const fw_rules_ipv4 = `*filter
:INPUT DROP [0:0]
:FORWARD DROP [0:0]
:OUTPUT DROP [0:0]

# Accept any related or established connections.
-I INPUT 1 -m state --state RELATED,ESTABLISHED -j ACCEPT
-I FORWARD 1 -m state --state RELATED,ESTABLISHED -j ACCEPT
-I OUTPUT 1 -m state --state {{ if not .Paranoid }}NEW,{{ end }}RELATED,ESTABLISHED -j ACCEPT

# Allow all traffic on the loopback interface.
-A INPUT -i lo -j ACCEPT
-A OUTPUT -o lo -j ACCEPT

{{ if .WithVPN }}
# Allow all traffic on the VPN interface.
-A INPUT -i tun0 -j ACCEPT
-A OUTPUT -o tun0 -j ACCEPT
{{ end }}

{{ if .Paranoid }}
# Outbound DNS lookups
-A OUTPUT -o {{ .Interface }} -p udp -m udp --dport 53 -j ACCEPT

# Outbound PING requests
-A OUTPUT -p icmp -j ACCEPT

# Outbound Network Time Protocol (NTP) request
-A OUTPUT -p udp --dport 123 --sport 123 -j ACCEPT

# Allow outbound DHCP request - Some hosts (Linode) automatically assign the primary IP
-A OUTPUT -p udp --dport 67:68 --sport 67:68 -m state --state NEW -j ACCEPT

# Outbound HTTP
-A OUTPUT -o {{ .Interface }} -p tcp -m tcp --dport 80 -m state --state NEW -j ACCEPT
-A OUTPUT -o {{ .Interface }} -p tcp -m tcp --dport 443 -m state --state NEW -j ACCEPT
{{ end }}

# SSH
-A INPUT -i {{ .Interface }} -p tcp -m tcp --dport 22 -m state --state NEW -j ACCEPT

{{ range .Rules }}{{ .Filter }}{{ end }}

{{ if .WithVPN }}
# Outbound OpenVPN traffic (required to connect to the VPN).
-A OUTPUT -o {{ .Interface }} -p tcp -m tcp --dport 1194 -m state --state NEW -j ACCEPT
-A OUTPUT -o {{ .Interface }} -p udp -m udp --dport 1194 -m state --state NEW -j ACCEPT
{{ end }}
COMMIT

*nat
{{ range .Rules }}{{ .NAT }}{{ end }}
COMMIT
`

const fwIpset = `{{ range .IPSets }}{{ .IPSetRestore }}{{ end }}`
const firewallUpstart = `#!/bin/sh
set -e

case "$IFACE" in
	{{ .Interface }})
		/usr/sbin/ipset restore -! < /etc/iptables/ipsets
		/sbin/iptables-restore < /etc/iptables/rules_ipv4
		/sbin/ip6tables-restore < /etc/iptables/rules_ipv6
		;;
esac

`
