package base

import (
	"fmt"
	. "github.com/dynport/zwo/cmd"
	"github.com/dynport/zwo/zwo"
	"strings"
)

type FWService struct {
	Description string
	Chain       string
	Port        int
	Interface   string
	Protocols   string
	Hosts       []string
}

type Firewall struct {
	PrimaryInterface string
	Services         []*FWService
	WithVPN          bool
	WithDHCP         bool
	Paranoid         bool
}

func (fw *Firewall) Compile(r *zwo.Runlist) {
	r.Execute(InstallPackages("iptables"))
	r.AddFile("/etc/iptables/rules_ipv4", "fw_rules_ipv4.conf", "root", 0644)
	r.AddFile("/etc/iptables/rules_ipv6", "fw_rules_ipv6.conf", "root", 0644)
	r.AddFile("/etc/network/if-pre-up.d/iptables", "fw_upstart.sh", "root", 0744)
	r.Execute("IFACE={{ .PrimaryInterface }} /etc/network/if-pre-up.d/iptables")
}

func (service *FWService) getInterface() string {
	iface := service.Interface
	if iface == "" {
		iface = "eth0"
	}
	return iface
}

func (service *FWService) getProtocols() []string {
	protosRaw := service.Protocols
	if protosRaw == "" {
		protosRaw = "tcp"
	}
	protos := strings.Split(protosRaw, ",")
	if len(protos) == 0 {
		protos = append(protos, "eth0")
	}
	return protos
}

func (fw *Firewall) GetServices() (rules string, e error) {
	for _, service := range fw.Services {
		if service.Chain != "INPUT" && service.Chain != "OUTPUT" {
			return "", fmt.Errorf("unsupported chain type '%s'", service.Chain)
		}

		iface := service.Interface
		if iface == "" {
			iface = fw.PrimaryInterface
		}

		rules += fmt.Sprintf("# %s %s\n", service.Chain, service.Description)
		for _, proto := range service.getProtocols() {
			rules += createRule(service.Chain, iface, proto, service.Port, service.Hosts)
		}
	}
	return rules, nil
}

func createRule(direction, iface, proto string, port int, hosts []string) string {
	prefix := ""
	sources, destinations := "", ""
	switch direction {
	case "INPUT":
		prefix = fmt.Sprintf("-A INPUT -i %s", iface)
		if len(hosts) > 0 {
			sources = "-s " + strings.Join(hosts, ",")
		}
	case "OUTPUT":
		prefix = fmt.Sprintf("-A OUTPUT -o %s", iface)
		if len(hosts) > 0 {
			destinations = "-d " + strings.Join(hosts, ",")
		}
	default:
		panic("should never ever happen")
	}
	return fmt.Sprintf("%s -p %s %s %s -m %s --dport %d -m state --state NEW -j ACCEPT\n", prefix, proto, sources, destinations, proto, port)
}
