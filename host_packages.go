package urknall

import (
	"github.com/dynport/dgtk/goup"
	. "github.com/dynport/urknall/cmd"
)

func runlist(name string, pkg Package) *Runlist {
	return &Runlist{name: name, pkg: pkg}
}

func (h *Host) buildSystemRunlists() {
	if h.Hostname != "" {
		h.addSystemPackage("hostname",
			h.newHostPackage(
				"hostname localhost", // Set hostname to make sudo happy.
				&FileCommand{Path: "/etc/hostname", Content: h.Hostname},
				&FileCommand{Path: "/etc/hosts", Content: "127.0.0.1 {{ .Hostname }} localhost"},
				"hostname -F /etc/hostname"))
	}

	if len(h.Firewall) > 0 {
		h.addSystemPackage("firewall",
			h.newHostPackage(
				InstallPackages("iptables", "ipset"),
				WriteAsset("/etc/network/if-pre-up.d/iptables", "fw_upstart.sh", "root", 0744),
				WriteAsset("/etc/iptables/ipsets", "fw_ipset.conf", "root", 0644),
				WriteAsset("/etc/iptables/rules_ipv4", "fw_rules_ipv4.conf", "root", 0644),
				WriteAsset("/etc/iptables/rules_ipv6", "fw_rules_ipv6.conf", "root", 0644),
				"{ modprobe iptable_filter && modprobe iptable_nat; }; /bin/true", // here to make sure next command succeeds.
				"IFACE={{ .Interface }} /etc/network/if-pre-up.d/iptables"))
	}

	if h.isDockerHost() {
		h.addSystemPackage("docker", &dockerPackage{Host: h})
	}
}

type hostPackage struct {
	*Host
	cmds []interface{}
}

func (h *Host) newHostPackage(cmds ...interface{}) *hostPackage {
	return &hostPackage{Host: h, cmds: cmds}
}

func (h *hostPackage) IsDockerHost() bool {
	return h.isDockerHost()
}

func (h *hostPackage) IsDockerBuildHost() bool {
	return h.isDockerHost()
}

func (h *hostPackage) Interface() string {
	return h.publicInterface()
}

func (hp *hostPackage) Package(rl *Runlist) {
	for i := range hp.cmds {
		rl.Add(hp.cmds[i])
	}
}

type dockerPackage struct {
	*Host
}

func (dp *dockerPackage) Package(rl *Runlist) {
	rl.Add(
		Or("grep universe /etc/apt/sources.list",
			And("sed 's/main$/main universe/' -i /etc/apt/sources.list",
				"DEBIAN_FRONTEND=noninteractive apt-get update")))
	rl.Add(
		InstallPackages("curl", "build-essential", "git-core", "bsdtar", "lxc", "aufs-tools"))
	rl.Add(
		Or(
			installDockerKernelOnRaring(),
			installDockerKernelOnPrecise(),
			"exit 1"))

	rl.Add(dp.dockerBinary())
	rl.Add(&UpstartCommand{Upstart: dp.createUpstart()})
	rl.Add("start docker")

	if dp.Docker.WithRegistry {
		rl.Add(WaitForUnixSocket("/var/run/docker.sock", 10))
		rl.Add("docker run -d -p 0.0.0.0:5000:5000 stackbrew/registry")
	}
}

func installDockerKernelOnRaring() *ShellCommand {
	return And("lsb_release -c | grep raring",
		InstallPackages("linux-image-extra-$(uname -r)"))
}

func installDockerKernelOnPrecise() *ShellCommand {
	return And("lsb_release -c | grep precise",
		IfNot("-f /etc/apt/sources.list.d/precise-updates.list",
			And("echo 'deb http://archive.ubuntu.com/ubuntu precise-updates main' > /etc/apt/sources.list.d/precise-updates.list",
				"apt-get update -y")),
		"apt-get -o Dpkg::Options::='--force-confnew' install linux-generic-lts-raring -y")
}

func (dp *dockerPackage) dockerBinary() *DownloadCommand {
	baseUrl := "http://get.docker.io/builds/Linux/x86_64"

	if dp.dockerVersion() < "0.6.0" {
		panic("version lower than 0.6.0 not supported yet")
	}
	url := baseUrl + "/docker-" + dp.dockerVersion()
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
