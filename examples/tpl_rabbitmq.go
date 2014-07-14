package main

import "github.com/dynport/urknall"

type RabbitMQ struct {
	Version string `urknall:"required=true"` // e.g. 3.3.4
}

func (p *RabbitMQ) Render(r urknall.Package) {
	r.AddCommands("packages",
		InstallPackages("erlang-nox", "erlang-reltool", "erlang-dev"),
	)
	r.AddCommands("download",
		DownloadAndExtract("{{ .Url }}", "/opt/"),
	)
	r.AddCommands("enable_management",
		Shell("cd {{ .InstallDir }} && ./sbin/rabbitmq-plugins enable rabbitmq_management"),
	)
	r.AddCommands("config",
		WriteFile("/etc/init/rabbitmq.conf", "env HOME=/root\nexec {{ .InstallDir }}/sbin/rabbitmq-server\n", "root", 0644),
	)
}

func (p *RabbitMQ) InstallDir() string {
	if p.Version == "" {
		panic("Version must be set")
	}
	return "/opt/rabbitmq_server-" + p.Version
}

func (p *RabbitMQ) Url() string {
	return "http://www.rabbitmq.com/releases/rabbitmq-server/v{{ .Version }}/rabbitmq-server-generic-unix-{{ .Version }}.tar.gz"
}
