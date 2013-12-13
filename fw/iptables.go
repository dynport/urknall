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
func (r *Rule) IPtablesRestore() (cmd string) {
	cfg := &iptConfig{rule: r, moduleConfig: map[string]iptModConfig{}}

	if r.Source != nil {
		r.Source.convert(cfg, "src")
	}

	if r.Destination != nil {
		r.Destination.convert(cfg, "dest")
	}

	return cfg.String()
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

	moduleConfig map[string]iptModConfig
}

func (cfg *iptConfig) String() (s string) {
	if cfg.rule.Description != "" {
		s = "# " + cfg.rule.Description + "\n"
	}
	s += "-A " + cfg.rule.Chain
	if cfg.sourceIP != "" {
		s += " --source " + cfg.sourceIP
	}
	if cfg.rule.Protocol != "" {
		s += " --protocol " + cfg.rule.Protocol
	}
	if cfg.sourceIface != "" {
		if cfg.rule.Chain == "FORWARD" {
			s += " --in-interface " + cfg.sourceIface
		} else {
			s += " --out-interface " + cfg.sourceIface
		}
	}
	if cfg.destIP != "" {
		s += " --destination " + cfg.destIP
	}
	if cfg.destIface != "" {
		if cfg.rule.Chain == "FORWARD" {
			s += " --out-interface " + cfg.destIface
		} else {
			s += " --in-interface " + cfg.destIface
		}
	}

	for module, modOptions := range cfg.moduleConfig {
		s += " " + module
		for option, value := range modOptions {
			s += " " + option + " " + value
		}
	}

	s += " -m state --state NEW -j ACCEPT\n"
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
}
