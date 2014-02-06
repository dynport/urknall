package docker

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

func New(version string) *Package {
	return &Package{Version: version}
}

type Package struct {
	Version   string `urknall:"default=0.7.3"`
	DataDir   string
	Public    bool
	Autostart bool
}

func (docker *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.InstallPackages("bsdtar", "lxc"),
		cmd.DownloadToFile("http://get.docker.io/builds/Linux/x86_64/docker-{{ .Version }}", "/opt/docker-{{ .Version }}", "root", 0755),
		cmd.WriteFile("/root/.dockercfg", "auth = abcbdne\nemail = a", "root", 0644),
		"ln -nfs /opt/docker-{{ .Version }} /usr/local/bin/docker",
		cmd.WriteFile("/etc/init/docker.conf", upstart, "root", 0644),
	)
	if docker.Autostart {
		r.Add("{ status docker | grep running; } || start docker") // no restarts for now
	}
}

const upstart = `
{{ if .Autostart }}
start on runlevel [2345]
stop on runlevel [!2345]
{{ end }}

exec /usr/local/bin/docker -d -D -H unix:///var/run/docker.sock -H tcp://{{ .NetworkInterface }}:4243 {{ with .DataDir }}-g={{ . }} 2>&1 | logger -i -t docker
`

func (docker *Package) NetworkInterface() string {
	if docker.Public {
		return "0.0.0.0"
	}
	return "127.0.0.1"
}
