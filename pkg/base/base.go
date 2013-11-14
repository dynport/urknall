// The BasePackage allows to do the basic host setup during provisioning.
package base

import (
	"github.com/dynport/zwo/assets"
	"github.com/dynport/zwo/templates"
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
func (base *BasePackage) Compile(r *zwo.Runlist) (e error) {
	if e = base.updateAndInstallPackages(r); e != nil {
		return e
	}

	if base.Hostname != "" {
		if e = base.setHostname(r); e != nil {
			return e
		}
	}

	if base.TimezoneUTC {
		if e = base.setTimezone(r); e != nil {
			return e
		}
	}

	if base.SyslogHost != "" {
		if e = base.setSyslogHost(r); e != nil {
			return e
		}
	}

	if base.SwapSize != "" {
		if e = base.setSwapFile(r); e != nil {
			return e
		}
	}

	if base.Limits {
		if e = base.setSysLimits(r); e != nil {
			return e
		}
	}

	if base.PublicIp != "" {
		if e = base.setNetwork(r); e != nil {
			return e
		}
	}
	return nil
}

func (base *BasePackage) updateAndInstallPackages(r *zwo.Runlist) (e error) {
	e = r.AddCommands(
		zwo.And(
			zwo.Execute("apt-get update"),
			zwo.Execute("DEBIAN_FRONTEND=noninteractive apt-get upgrade -y")))
	if e != nil {
		return e
	}
	if len(base.Packages) > 0 {
		return r.AddCommands(zwo.InstallPackages(base.Packages...))
	}
	return nil
}

func (base *BasePackage) setHostname(r *zwo.Runlist) (e error) {
	e = r.AddFiles(
		zwo.WriteFile("/etc/hostname", base.Hostname, "root", 0755),
		zwo.WriteFile("/etc/hosts", "127.0.0.1 {{ .Hostname }} localhost", "root", 0755))
	if e != nil {
		return e
	}
	return r.AddCommands(zwo.Execute("hostname -F /etc/hostname"))
}

func (base *BasePackage) setTimezone(r *zwo.Runlist) (e error) {
	return r.AddCommands(
		zwo.Execute(`echo "Etc/UTC" | tee /etc/timezone && sudo dpkg-reconfigure --frontend noninteractive tzdata`))
}

func (base *BasePackage) setSyslogHost(r *zwo.Runlist) (e error) {
	cfgContent, e := templates.RenderAssetFromString("syslog.conf", base)
	if e != nil {
		return e
	}
	e = r.AddCommands(zwo.InstallPackages("rsyslog"))
	if e != nil {
		return e
	}
	e = r.AddFiles(zwo.WriteFile("/etc/rsyslog.d/50-default.conf", cfgContent, "root", 0600))
	if e != nil {
		return e
	}
	return r.AddCommands(zwo.Execute("/etc/init.d/rsyslog restart"))
}

func (base *BasePackage) setSwapFile(r *zwo.Runlist) (e error) {
	return r.AddCommands(
		zwo.And(
			zwo.Execute("dd if=/dev/zero of=/swapfile bs=1024 count={{ .SwapSize }}k"),
			zwo.Execute("mkswap /swapfile"),
			zwo.Execute("swapon /swapfile")))
}

func (base *BasePackage) setSysLimits(r *zwo.Runlist) (e error) {
	limitCfgBytes, e := assets.Get("limits.conf")
	if e != nil {
		return e
	}
	limitCfg := string(limitCfgBytes)

	sysctlCfg, e := templates.RenderAssetFromString("sysctl.conf", base)
	if e != nil {
		return e
	}

	e = r.AddFiles(
		zwo.WriteFile("/etc/security/limits.conf", limitCfg, "root", 0600),
		zwo.WriteFile("/etc/sysctl.conf", sysctlCfg, "root", 0600))
	if e != nil {
		return e
	}

	return r.AddCommands(
		zwo.And(
			zwo.Execute("ulimit -a"),
			zwo.Execute("sysctl -p")))
}

func (base *BasePackage) setNetwork(r *zwo.Runlist) (e error) {
	networkCfg, e := templates.RenderAssetFromString("network.cfg", base)
	if e != nil {
		return e
	}
	return r.AddFiles(zwo.WriteFile("/etc/network/interfaces", networkCfg, "root", 0600))
}
