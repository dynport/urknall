package rabbitmq

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

func New(version string) *Package {
	return &Package{Version: version}
}

type Package struct {
	Version string `urknall:"default="3.2.1"`
}

func (p *Package) Package(r *urknall.Package) {
	r.Add(
		cmd.InstallPackages("erlang-nox", "erlang-reltool", "erlang-dev"),
		cmd.Mkdir("/opt/src/", "root", 0755),
		cmd.DownloadAndExtract(p.url(), "/opt/"),
		"cd {{ .InstallPath }} && ./sbin/rabbitmq-plugins enable rabbitmq_management",
		cmd.WriteFile("/etc/init/rabbitmq.conf", "env HOME=/root\nexec {{ .InstallPath }}/sbin/rabbitmq-server\n", "root", 0644),
	)
}

func (p *Package) InstallPath() string {
	return "/opt/rabbitmq_server-{{ .Version }}"
}

func (p *Package) url() string {
	return "http://www.rabbitmq.com/releases/rabbitmq-server/v{{ .Version }}/rabbitmq-server-generic-unix-{{ .Version }}.tar.gz"
}
