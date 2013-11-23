package docker

import (
	"github.com/dynport/dgtk/goup"
	. "github.com/dynport/zwo/cmd"
	"github.com/dynport/zwo/zwo"
)

type Host struct {
	Version      string
	Public       bool
	Debug        bool
	WithRegistry bool
}

func (d *Host) Compile(r *zwo.Runlist) {
	r.Execute(
		Or("grep universe /etc/apt/sourceslist",
			And("sed 's/main$/main universe/' -i /etc/apt/sources.list",
				"apt-get update")))
	r.Execute(
		InstallPackages("curl", "build-essential", "git-core", "bsdtar", "lxc", "aufs-tools"))
	r.Execute(
		Or(
			installDockerKernelOnRaring(),
			installDockerKernelOnPrecise(),
			"exit 1"))

	r.Execute(d.getDockerBinary())
	r.Init(d.createUpstart(), "")
	r.Execute("start docker")

	if d.WithRegistry {
		r.RunDockerImage("stackbrew/registry", "", 5000)
		// r.Execute("docker run -d -p 0.0.0.0:5000:5000 stackbrew/registry")
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

func (d *Host) getDockerBinary() string {
	baseUrl := "http://get.docker.io/builds/Linux/x86_64"

	if d.Version < "0.6.0" {
		panic("version lower than 0.6.0 not supported yet")
	}
	url := baseUrl + "/docker-" + d.Version
	return DownloadToFile(url, "/usr/local/bin/docker", "root", 0700)
}

func (d *Host) createUpstart() *goup.Upstart {
	execCmd := "/usr/local/bin/docker -d -r -H unix:///var/run/docker.sock"
	if d.Debug {
		execCmd += " -D"
	}
	if d.Public {
		execCmd += " -H tcp://0.0.0.0:4243"
	}
	execCmd += " 2>&1 | logger -i -t docker"
	return &goup.Upstart{
		Name:          "docker",
		StartOnEvents: []string{"runlevel [2345]"},
		StopOnEvents:  []string{"runlevel [!2345]"},
		Exec:          execCmd,
	}
}
