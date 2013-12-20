package metrix

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/utils"
)

func New(version string) *Package {
	return &Package{Version: version}
}

type Package struct {
	Version         string `urknall:"default=0.1.6"`
	Hostname        string
	OpentsdbAddress string
	AmqpAddress     string
	NginxUrl        string
	LoadAvg         bool
	Memory          bool
	Cpu             bool
	Disk            bool
	Processes       bool
	Net             bool
	Df              bool
	Free            bool
	ElasticSearch   string
}

func (metric *Package) url() string {
	return "https://github.com/dynport/metrix/releases/download/v{{ .Version }}/metrix-v{{ .Version }}.linux.amd64.tar.gz"
}

func (metrix *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.Mkdir(metrix.installPath(), "root", 0755),
		cmd.DownloadAndExtract(metrix.url(), metrix.installPath()),
		cmd.WriteFile("/etc/cron.d/metrix", "* * * * * root "+metrix.cmd()+" \n", "root", 0644),
	)
	return
}

func (metrix *Package) installPath() string {
	return "/opt/metrix-{{ .Version }}"
}

func (metrix *Package) cmd() string {
	cmd := metrix.installPath() + "/metrix"
	if metrix.LoadAvg {
		cmd += " --loadavg"
	}
	if metrix.Memory {
		cmd += " --memory"
	}
	if metrix.Cpu {
		cmd += " --cpu"
	}
	if metrix.Disk {
		cmd += " --disk"
	}
	if metrix.Processes {
		cmd += " --processes"
	}
	if metrix.Net {
		cmd += " --net"
	}
	if metrix.Df {
		cmd += " --df"
	}
	if metrix.Free {
		cmd += " --free"
	}
	if metrix.OpentsdbAddress != "" {
		cmd += " --opentsdb=" + metrix.OpentsdbAddress
	}

	if metrix.ElasticSearch != "" {
		cmd += " --elasticsearch=" + metrix.ElasticSearch
	}

	minAmqpVersion, e := utils.ParseVersion("0.1.5")
	if e != nil {
		panic("unable to parse minAmqpVersion")
	}

	v, e := utils.ParseVersion(metrix.Version)
	if e != nil {
		panic("Error parsing version: " + e.Error())
	}

	if metrix.AmqpAddress != "" {
		if v.Smaller(minAmqpVersion) {
			panic("amqp requires at least version " + minAmqpVersion.String())
		}
		cmd += " --amqp=" + metrix.AmqpAddress
	}

	if metrix.NginxUrl != "" {
		cmd += " --nginx=" + metrix.NginxUrl
	}

	return cmd + " 2>&1 | logger -i -t metrix"
}
