// The BasePackage allows to do the basic host setup during provisioning.
package base

import (
	. "github.com/dynport/zwo/cmd"
	"github.com/dynport/zwo/zwo"
)

// The packages configuration.
type BasePackage struct {
	Hostname    string `json:"hostname"`
	PublicIp    string `json:"public_ip"`
	TimezoneUTC bool   `json:"utc"`
	SyslogHost  string `json:"syslog_host"`
	SwapSize    string `json:"swap"`
	Limits      bool   `json:"limits"`
	ShmMax      string `json:"shmmax"`
	ShmAll      string `json:"shmall"`
	Packages    []string
}

// Stops on all errors in the process of creating the runlist (like insufficient configuration for example).
func (base *BasePackage) Compile(r *zwo.Runlist) {
	base.updateAndInstallPackages(r)

	if base.Hostname != "" {
		base.setHostname(r)
	}

	if base.TimezoneUTC {
		base.setTimezone(r)
	}

	if base.SyslogHost != "" {
		base.setSyslogHost(r)
	}

	if base.SwapSize != "" {
		base.setSwapFile(r)
	}

	if base.Limits {
		base.setSysLimits(r)
	}

	if base.PublicIp != "" {
		base.setNetwork(r)
	}
}

func (base *BasePackage) updateAndInstallPackages(r *zwo.Runlist) {
	r.Execute(
		And("apt-get update",
			"DEBIAN_FRONTEND=noninteractive apt-get upgrade -y"))
	if len(base.Packages) > 0 {
		r.Execute(InstallPackages(base.Packages...))
	}
}

func (base *BasePackage) setHostname(r *zwo.Runlist) {
	r.AddFile("/etc/hostname", base.Hostname, "root", 0755)
	r.AddFile("/etc/hosts", "127.0.0.1 {{ .Hostname }} localhost", "root", 0755)
	r.Execute("hostname -F /etc/hostname")
}

func (base *BasePackage) setTimezone(r *zwo.Runlist) {
	r.Execute(
		And("echo 'Etc/UTC' | tee /etc/timezone",
			"sudo dpkg-reconfigure --frontend noninteractive tzdata"))
}

func (base *BasePackage) setSyslogHost(r *zwo.Runlist) {
	r.Execute(InstallPackages("rsyslog"))
	r.AddFile("/etc/rsyslog.d/50-default.conf", "syslog.conf", "root", 0600)
	r.Execute("/etc/init.d/rsyslog restart")
}

func (base *BasePackage) setSwapFile(r *zwo.Runlist) {
	r.Execute(
		And("dd if=/dev/zero of=/swapfile bs=1024 count={{ .SwapSize }}k",
			"mkswap /swapfile",
			"swapon /swapfile"))
}

func (base *BasePackage) setSysLimits(r *zwo.Runlist) {
	r.AddFile("/etc/security/limits.conf", "limits.conf", "root", 0600)
	r.AddFile("/etc/sysctl.conf", "sysctl.conf", "root", 0600)
	r.Execute(
		And("ulimit -a",
			"sysctl -p"))
}

func (base *BasePackage) setNetwork(r *zwo.Runlist) {
	r.AddFile("/etc/network/interfaces", "network.cfg", "root", 0600)
}
