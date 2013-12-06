package zwo

import (
	"github.com/dynport/dgtk/goup"
	. "github.com/dynport/zwo/cmd"
	"github.com/dynport/zwo/host"
)

func createHostPackages(host *host.Host) (p []Compiler) {
	p = []Compiler{}
	p = append(p, &hostPackage{Host: host})
	p = append(p, &firewallPackage{Host: host})

	if host.IsDockerHost() {
		p = append(p, &dockerPackage{Host: host})
	}

	return p
}

type hostPackage struct {
	*host.Host
}

func (hp *hostPackage) Compile(rl *Runlist) {
	if rl.host.Hostname() != "" { // Set hostname.
		rl.AddFile("/etc/hostname", hp.Hostname(), "root", 0755)
		rl.AddFile("/etc/hosts", "127.0.0.1 {{ .Hostname }} localhost", "root", 0755)
		rl.Execute("hostname -F /etc/hostname")
	}
}

func (hp *hostPackage) CompileName() string {
	return "zwo.host"
}


type firewallPackage struct {
	*host.Host
}

func (fw *firewallPackage) Compile(rl *Runlist) {
	rl.Execute(InstallPackages("iptables", "ipset"))

	rl.AddAsset("/etc/network/if-pre-up.d/iptables", "fw_upstart.sh", "root", 0744)
	rl.AddAsset("/etc/iptables/rules_ipv4", "fw_rules_ipv4.conf", "root", 0644)
	rl.AddAsset("/etc/iptables/rules_ipv6", "fw_rules_ipv6.conf", "root", 0644)
	rl.Execute("modprobe iptable_filter && modprobe iptable_nat") // here to make sure next command succeeds.
	rl.Execute("IFACE={{ .Interface }} /etc/network/if-pre-up.d/iptables")
}

func (fw *firewallPackage) CompileName() string {
	return "zwo.fw"
}

type dockerPackage struct {
	*host.Host
}

func (dp *dockerPackage) CompileName() string {
	return "zwo.docker"
}

func (dp *dockerPackage) Compile(rl *Runlist) {
	rl.Execute(
		Or("grep universe /etc/apt/sourceslist",
			And("sed 's/main$/main universe/' -i /etc/apt/sources.list",
				"apt-get update")))
	rl.Execute(
		InstallPackages("curl", "build-essential", "git-core", "bsdtar", "lxc", "aufs-tools"))
	rl.Execute(
		Or(
			installDockerKernelOnRaring(),
			installDockerKernelOnPrecise(),
			"exit 1"))

	rl.Execute(dp.getDockerBinary())
	rl.Init(dp.createUpstart(), "")
	rl.Execute("start docker")

	if dp.Docker.WithRegistry {
		rl.Execute(
			And(WaitForUnixSocket("/var/run/docker.sock", 10),
				"docker run -d -p 0.0.0.0:5000:5000 stackbrew/registry"))
	}
}

func installDockerKernelOnRaring() string {
	return And("lsb_release -c | grep raring",
		InstallPackages("linux-image-extra-$(uname -r)"))
}

func installDockerKernelOnPrecise() string {
	return And("lsb_release -c | grep precise",
		IfNot("-f /etc/apt/sources.list.d/precise-updates.list",
			And("echo 'deb http://archive.ubuntu.com/ubuntu precise-updates main' > /etc/apt/sources.list.d/precise-updates.list",
				"apt-get update -y")),
		"apt-get -o Dpkg::Options::='--force-confnew' install linux-generic-lts-raring -y")
}

func (dp *dockerPackage) getDockerBinary() string {
	baseUrl := "http://get.docker.io/builds/Linux/x86_64"

	if dp.DockerVersion() < "0.6.0" {
		panic("version lower than 0.6.0 not supported yet")
	}
	url := baseUrl + "/docker-" + dp.DockerVersion()
	return DownloadToFile(url, "/usr/local/bin/docker", "root", 0700)
}

func (dp *dockerPackage) createUpstart() *goup.Upstart {
	execCmd := "/usr/local/bin/docker -d -r -H unix:///var/run/docker.sock -H tcp://127.0.0.1:4243 2>&1 | logger -i -t docker"
	return &goup.Upstart{
		Name:          "docker",
		StartOnEvents: []string{"runlevel [2345]"},
		StopOnEvents:  []string{"runlevel [!2345]"},
		Exec:          execCmd,
	}
}
