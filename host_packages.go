package urknall

import (
	"github.com/dynport/urknall/cmd"
)

func newRunlist(name string, pkg Packager, host *Host) *Runlist {
	return &Runlist{name: name, pkg: pkg, host: host}
}

func (h *Host) buildSystemRunlists() {
	if h.Hostname != "" {
		h.addSystemPackage("hostname",
			h.newHostPackage(
				"hostname localhost", // Set hostname to make sudo happy.
				&cmd.FileCommand{Path: "/etc/hostname", Content: h.Hostname},
				&cmd.FileCommand{Path: "/etc/hosts", Content: "127.0.0.1 {{ .Hostname }} localhost"},
				"hostname -F /etc/hostname"))
	}

	if h.Timezone != "" {
		h.addSystemPackage("timezone",
			h.newHostPackage(
				cmd.WriteFile("/etc/timezone", h.Timezone, "root", 0644),
				"dpkg-reconfigure --frontend noninteractive tzdata",
			),
		)
	}

	if len(h.Firewall) > 0 {
		h.addSystemPackage("firewall",
			h.newHostPackage(
				cmd.InstallPackages("iptables", "ipset"),
				cmd.WriteFile("/etc/network/if-pre-up.d/iptables", string(mustReadAsset("fw_upstart.sh")), "root", 0744),
				cmd.WriteFile("/etc/iptables/ipsets", string(mustReadAsset("fw_ipset.conf")), "root", 0644),
				cmd.WriteFile("/etc/iptables/rules_ipv4", string(mustReadAsset("fw_rules_ipv4.conf")), "root", 0644),
				cmd.WriteFile("/etc/iptables/rules_ipv6", string(mustReadAsset("fw_rules_ipv6.conf")), "root", 0644),
				"{ modprobe iptable_filter && modprobe iptable_nat; }; /bin/true", // here to make sure next command succeeds.
				"IFACE={{ .Interface }} /etc/network/if-pre-up.d/iptables"))
	}
}

type hostPackage struct {
	*Host
	cmds []interface{}
}

func (h *Host) newHostPackage(cmds ...interface{}) *hostPackage {
	return &hostPackage{Host: h, cmds: cmds}
}

func (h *hostPackage) Interface() string {
	return h.publicInterface()
}

func (hp *hostPackage) Package(rl *Runlist) {
	for i := range hp.cmds {
		rl.Add(hp.cmds[i])
	}
}
