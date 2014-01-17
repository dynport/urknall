package openvpn

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type Masquerade struct {
	Interface string `urknall:"required=true"`
}

func (*Masquerade) Package(r *urknall.Runlist) {
	r.Add(
		cmd.WriteFile("/etc/network/if-pre-up.d/iptables", ipUp, "root", 0744),
		"IFACE={{ .Interface }} /etc/network/if-pre-up.d/iptables",
	)
}

const ipUp = `#!/bin/bash -e

if [[ "$IFACE" == "{{ .Interface }}" ]]; then
	echo 1 > /proc/sys/net/ipv4/ip_forward
	iptables -t nat -A POSTROUTING -o {{ .Interface }} -j MASQUERADE
fi
`
