package fw

import (
	"net"
	"strconv"
)

// The target of a rule. It can be specified either by IP or the name of an IPSet. Additional parameters are the port
// and interface used. It's totally valid to only specify a subset (or even none) of the fields. For example IP and
// IPSet must not be given for the host the rule is applied on.
//
// TODO(gfrey): There currently is no validation the referenced IPSet exists. This should be added on provisioning to
// make sure iptables setup won't fail.
type Target struct {
	IP        net.IP // IP of the target.
	IPSet     string // IPSet used for matching.
	Port      int    // Port packets must use to match.
	Interface string // Interface the packet goes through.
	NAT       string // NAT configuration (empty, "MASQ", or Interface's IP).
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
type Rule struct {
	Description string
	Chain       string // Chain to add the rule to.
	Protocol    string // The protocol used.

	Source      *Target // The source of the packet.
	Destination *Target // The destination of the packet.
}

// Method to create something digestable for IPtablesRestore (aka users might well ignore this).
func (r *Rule) Filter() (cmd string) {
	cfg := &iptConfig{rule: r, moduleConfig: map[string]iptModConfig{}}

	if r.Source != nil {
		r.Source.convert(cfg, "src")
	}

	if r.Destination != nil {
		r.Destination.convert(cfg, "dest")
	}

	return cfg.FilterTableRule()
}

func (r *Rule) isNATRule() bool {
	return r.Chain == "FORWARD" &&
		((r.Source != nil && r.Source.NAT != "") ||
			(r.Destination != nil && r.Destination.NAT != ""))
}

func (r *Rule) NAT() (cmd string) {
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

func (r *Rule) IPsets() {
}

type iptModConfig map[string]string

type iptConfig struct {
	rule *Rule

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

func (t *Target) convert(cfg *iptConfig, tType string) {
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
