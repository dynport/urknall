package main

import "github.com/dynport/urknall"

type RabbitMQ struct {
	Version string `urknall:"required=true"` // e.g. 3.3.4
}

func (amqp *RabbitMQ) Render(pkg urknall.Package) {
	pkg.AddCommands("packages",
		InstallPackages("erlang-nox", "erlang-reltool", "erlang-dev"),
	)
	pkg.AddCommands("download",
		DownloadAndExtract("{{ .Url }}", "/opt/"),
	)
	pkg.AddCommands("enable_management",
		Shell("cd {{ .InstallDir }} && ./sbin/rabbitmq-plugins enable rabbitmq_management"),
	)
	pkg.AddCommands("config",
		WriteFile("/etc/init/rabbitmq.conf", "env HOME=/root\nexec {{ .InstallDir }}/sbin/rabbitmq-server\n", "root", 0644),
	)
}

func (amqp *RabbitMQ) InstallDir() string {
	if amqp.Version == "" {
		panic("Version must be set")
	}
	return "/opt/rabbitmq_server-" + amqp.Version
}

func (amqp *RabbitMQ) Url() string {
	return "http://www.rabbitmq.com/releases/rabbitmq-server/v{{ .Version }}/rabbitmq-server-generic-unix-{{ .Version }}.tar.gz"
}
