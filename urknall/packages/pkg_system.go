package main

import "github.com/dynport/urknall"

type Limits struct {
}

func (limits *Limits) Package(r *urknall.Package) {
	r.Add(
		WriteFile("/etc/security/limits.conf", limitsTpl, "root", 0644),
		"ulimit -a",
	)
}

const limitsTpl = `* soft nofile 65535
* hard nofile 65535
root soft nofile 65535
root hard nofile 65535
`

type SysCtl struct {
	ShmMax string
	ShmAll string
}

func (sysctl *SysCtl) Package(r *urknall.Package) {
	r.Add(
		WriteFile("/etc/sysctl.conf", sysctlTpl, "root", 0644),
		"sysctl -p",
	)
}

const sysctlTpl = `net.core.rmem_max=16777216
net.core.wmem_max=16777216
net.core.wmem_default=262144
net.ipv4.tcp_rmem=4096 87380 16777216
net.ipv4.tcp_wmem=4096 65536 16777216
net.core.netdev_max_backlog=4000
net.ipv4.tcp_low_latency=1
net.ipv4.tcp_window_scaling=1
net.ipv4.tcp_timestamps=1
net.ipv4.tcp_sack=1
fs.file-max=65535
net.core.wmem_default=8388608
net.core.rmem_default=8388608
net.core.netdev_max_backlog=10000
net.core.somaxconn=4000
net.ipv4.tcp_max_syn_backlog=40000
net.ipv4.tcp_fin_timeout=15
net.ipv4.tcp_tw_reuse=1
vm.swappiness=0
{{ if .ShmMax }}kernel.shmmax={{ .ShmMax }}{{ end }}
{{ if .ShmAll }}kernel.shmmax={{ .ShmAll }}{{ end }}
`

type Timezone struct {
	Timezone string `urknall:"required=true"`
}

func (t *Timezone) Package(r *urknall.Package) {
	r.Add(
		WriteFile("/etc/timezone", t.Timezone, "root", 0644),
		"dpkg-reconfigure --frontend noninteractive tzdata",
	)
}

type Hostname struct {
	Hostname string `urknall:"required=true"`
}

func (h *Hostname) Package(r *urknall.Package) {
	r.Add(
		"hostname localhost", // Set hostname to make sudo happy.
		&FileCommand{Path: "/etc/hostname", Content: h.Hostname},
		&FileCommand{Path: "/etc/hosts", Content: "127.0.0.1 {{ .Hostname }} localhost"},
		"hostname -F /etc/hostname",
	)
}
