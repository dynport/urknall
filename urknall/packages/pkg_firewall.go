package main

import (
	"fmt"
	"net"
	"strconv"

	"github.com/dynport/urknall"
)

type Firewall struct {
	Interface string `urknall:"default=eth0"`
	WithVPN   bool
	Paranoid  bool
	Rules     []*FirewallRule
	IPSets    []*FirewallIPSet // List of ipsets for the firewall.
}

func (f *Firewall) Render(r urknall.Package) {
	t := urknall.NewTask()
	t.SetCacheKey("base")
	t.Add(
		InstallPackages("iptables", "ipset"),
		WriteFile("/etc/network/if-pre-up.d/iptables", firewallUpstart, "root", 0744),
	)
	if len(f.IPSets) > 0 {
		t.Add(WriteFile("/etc/iptables/ipsets", fwIpset, "root", 0644))
	}
	t.Add(
		WriteFile("/etc/iptables/rules_ipv4", fw_rules_ipv4, "root", 0644),
		WriteFile("/etc/iptables/rules_ipv6", fw_rules_ipv6, "root", 0644),
		"{ modprobe iptable_filter && modprobe iptable_nat; }; /bin/true", // here to make sure next command succeeds.
		"IFACE={{ .Interface }} /etc/network/if-pre-up.d/iptables",
	)
	r.AddTask(t)
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

func (ips *FirewallIPSet) IPSetRestore() (cmd string) {
	cmd = fmt.Sprintf("create %s %s family %s hashsize %d maxelem %d\n",
		ips.Name, ips.Type, ips.family(), ips.hashsize(), ips.maxelem())
	for idx := range ips.Members {
		cmd += fmt.Sprintf("add %s %s\n", ips.Name, ips.Members[idx].String())
	}
	return cmd + "\n"
}

func (ips *FirewallIPSet) family() string {
	if ips.Family == "" {
		return "inet"
	}
	return ips.Family
}

func (i *FirewallIPSet) hashsize() int {
	if i.HashSize == 0 {
		return 1024
	}
	return i.HashSize
}

func (i *FirewallIPSet) maxelem() int {
	if i.MaxElem == 0 {
		return 65536
	}
	return i.MaxElem
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

// Method to create something digestable for IPtablesRestore (aka users might well ignore this).
func (r *FirewallRule) Filter() (cmd string) {
	cfg := &iptConfig{rule: r, moduleConfig: map[string]iptModConfig{}}

	if r.Source != nil {
		r.Source.convert(cfg, "src")
	}

	if r.Destination != nil {
		r.Destination.convert(cfg, "dest")
	}

	return cfg.FilterTableRule()
}

func (r *FirewallRule) isNATRule() bool {
	return r.Chain == "FORWARD" &&
		((r.Source != nil && r.Source.NAT != "") ||
			(r.Destination != nil && r.Destination.NAT != ""))
}

func (r *FirewallRule) NAT() (cmd string) {
	if !r.isNATRule() {
		return ""
	}

	cfg := &iptConfig{rule: r, moduleConfig: map[string]iptModConfig{}}

	if r.Source != nil {
		r.Source.convert(cfg, "src")
	}

	if r.Destination != nil {
		r.Destination.convert(cfg, "dest")
	}

	return cfg.NATTableRule()
}

func (r *FirewallRule) IPsets() {
}

type iptModConfig map[string]string

type iptConfig struct {
	rule *FirewallRule

	sourceIP string
	destIP   string

	sourceIface string
	destIface   string

	sourceNAT string
	destNAT   string

	moduleConfig map[string]iptModConfig
}

func (cfg *iptConfig) basicSettings(natRule bool) (s string) {
	if cfg.rule.Protocol != "" {
		s += " --protocol " + cfg.rule.Protocol
	}
	if cfg.sourceIP != "" {
		s += " --source " + cfg.sourceIP
	}
	if cfg.sourceIface != "" {
		if cfg.rule.Chain == "FORWARD" {
			if !natRule || cfg.destNAT != "" {
				s += " --in-interface " + cfg.sourceIface
			}
		} else {
			s += " --out-interface " + cfg.sourceIface
		}
	}
	if cfg.destIP != "" {
		s += " --destination " + cfg.destIP
	}
	if cfg.destIface != "" {
		if cfg.rule.Chain == "FORWARD" {
			if !natRule || cfg.sourceNAT != "" {
				s += " --out-interface " + cfg.destIface
			}
		} else {
			s += " --in-interface " + cfg.destIface
		}
	}
	return s
}

func (cfg *iptConfig) FilterTableRule() (s string) {
	if cfg.rule.Description != "" {
		s = "# " + cfg.rule.Description + "\n"
	}
	s += "-A " + cfg.rule.Chain

	s += cfg.basicSettings(false)

	for module, modOptions := range cfg.moduleConfig {
		s += " " + module
		for option, value := range modOptions {
			s += " " + option + " " + value
		}
	}

	s += " -m state --state NEW -j ACCEPT\n"
	return s
}

func (cfg *iptConfig) NATTableRule() (s string) {
	if cfg.rule.Description != "" {
		s = "# " + cfg.rule.Description + "\n"
	}

	switch {
	case cfg.sourceNAT != "" && cfg.destNAT == "":
		s += "-A POSTROUTING"
	case cfg.sourceNAT == "" && cfg.destNAT != "":
		s += "-A PREROUTING"
	default:
		panic("but you said NAT would be configured?!")
	}

	s += cfg.basicSettings(true)

	switch {
	case cfg.sourceNAT == "MASQ":
		s += " -j MASQUERADE"
	case cfg.sourceNAT != "":
		s += " -j SNAT --to " + cfg.sourceNAT
	case cfg.destNAT != "":
		s += " -j DNAT --to " + cfg.destNAT
	}

	return s
}

func (t *FirewallTarget) convert(cfg *iptConfig, tType string) {
	if t.Port != 0 {
		if cfg.rule.Protocol == "" {
			panic("port requires the protocol to be specified")
		}

		module := "-m " + cfg.rule.Protocol
		if _, found := cfg.moduleConfig[module]; !found {
			cfg.moduleConfig[module] = iptModConfig{}
		}
		switch tType {
		case "src":
			cfg.moduleConfig[module]["--source-port"] = strconv.Itoa(t.Port)
		case "dest":
			cfg.moduleConfig[module]["--destination-port"] = strconv.Itoa(t.Port)
		}
	}

	if t.IP != nil {
		switch tType {
		case "src":
			cfg.sourceIP = t.IP.String()
		case "dest":
			cfg.destIP = t.IP.String()
		}
	}

	if t.IPSet != "" {
		module := "-m set"
		if _, found := cfg.moduleConfig[module]; !found {
			cfg.moduleConfig[module] = iptModConfig{}
		}
		value := cfg.moduleConfig[module]["--match-set "+t.IPSet]
		set := ""
		switch tType {
		case "src":
			set = "src"
		case "dest":
			set = "dst"
		}
		if value != "" {
			cfg.moduleConfig[module]["--match-set "+t.IPSet] = value + "," + set
		} else {
			cfg.moduleConfig[module]["--match-set "+t.IPSet] = set
		}
	}

	if t.Interface != "" {
		switch tType {
		case "src":
			cfg.sourceIface = t.Interface
		case "dest":
			cfg.destIface = t.Interface
		}
	}

	if t.NAT != "" {
		switch tType {
		case "src": // for input on the source the destination address can be modified.
			cfg.destNAT = t.NAT
		case "dest": // for output on the destination the source address can be modified.
			cfg.sourceNAT = t.NAT
		}

		if cfg.sourceNAT != "" && cfg.destNAT != "" {
			panic("only source or destination NAT allowed!")
		}
	}
}

// The target of a rule. It can be specified either by IP or the name of an IPSet. Additional parameters are the port
// and interface used. It's totally valid to only specify a subset (or even none) of the fields. For example IP and
// IPSet must not be given for the host the rule is applied on.
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

const fwIpset = `# IPSet configuration
{{ range .IPSets }}{{ .IPSetRestore }}{{ end }}`
const firewallUpstart = `#!/bin/sh
set -e

case "$IFACE" in
	{{ .Interface }})
		test -e /etc/iptables/ipsets && /usr/sbin/ipset restore -! < /etc/iptables/ipsets
		/sbin/iptables-restore < /etc/iptables/rules_ipv4
		/sbin/ip6tables-restore < /etc/iptables/rules_ipv6
		;;
esac

`
