package zwo

import (
	"github.com/dynport/dgtk/goup"
	. "github.com/dynport/zwo/cmd"
	"github.com/dynport/zwo/firewall"
	"github.com/dynport/zwo/host"
)

func createHostPreRunlist(h *host.Host) (rl *Runlist) {
	rl = &Runlist{host: h}
	rl.setName("int:host.setup")
	rl.setConfig(h)

	if h.Hostname() != "" { // Set hostname.
		rl.AddFile("/etc/hostname", h.Hostname(), "root", 0755)
		rl.AddFile("/etc/hosts", "127.0.0.1 {{ .Hostname }} localhost", "root", 0755)
		rl.Execute("hostname -F /etc/hostname")
	}

	rl.Execute(InstallPackages("iptables", "ipset"))

	// Write initial set of iptables rules to make sure system is not open during installation.
	rl.AddAsset("/etc/network/if-pre-up.d/iptables", "fw_upstart.sh", "root", 0744)
	setupIPTables(rl)
	installDocker(h, rl)

	return rl
}

func createHostPostRunlist(h *host.Host) (rl *Runlist) {
	rl = &Runlist{host: h}
	rl.setName("int:host.firewall")
	rl.setConfig(h)
	setupIPTables(rl)

	return rl
}

func setupIPTables(rl *Runlist) {
	rl.AddAsset("/etc/iptables/rules_ipv4", "fw_rules_ipv4.conf", "root", 0644)
	rl.AddAsset("/etc/iptables/rules_ipv6", "fw_rules_ipv6.conf", "root", 0644)
	rl.Execute("modprobe iptable_filter && modprobe iptable_nat") // here to make sure next command succeeds.
	rl.Execute("IFACE={{ .Interface }} /etc/network/if-pre-up.d/iptables")
}

func installDocker(h *host.Host, rl *Runlist) {
	if !h.IsDockerHost() {
		return
	}

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

	rl.Execute(getDockerBinary(h))
	rl.Init(createUpstart(h), "")
	rl.Execute("start docker")

	if h.Docker.WithRegistry {
		rl.WaitForUnixSocket("/var/run/docker.sock", 10)
		rl.Execute("docker run -d -p 0.0.0.0:5000:5000 stackbrew/registry")
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

func getDockerBinary(h *host.Host) string {
	baseUrl := "http://get.docker.io/builds/Linux/x86_64"

	if h.DockerVersion() < "0.6.0" {
		panic("version lower than 0.6.0 not supported yet")
	}
	url := baseUrl + "/docker-" + h.DockerVersion()
	return DownloadToFile(url, "/usr/local/bin/docker", "root", 0700)
}

func createUpstart(h *host.Host) *goup.Upstart {
	execCmd := "/usr/local/bin/docker -d -r -H unix:///var/run/docker.sock -H tcp://127.0.0.1:4243 2>&1 | logger -i -t docker"
	return &goup.Upstart{
		Name:          "docker",
		StartOnEvents: []string{"runlevel [2345]"},
		StopOnEvents:  []string{"runlevel [!2345]"},
		Exec:          execCmd,
	}
}
