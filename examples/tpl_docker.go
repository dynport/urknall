package main

import "github.com/dynport/urknall"

type Docker struct {
	Version string `urknall:"required=true"`
}

func (tpl *Docker) Render(p urknall.Package) {
	p.AddCommands("packages", InstallPackages("aufs-tools", "cgroup-lite", "xz-utils", "git"))
	p.AddCommands("install",
		Mkdir("{{ .InstallDir }}/bin", "root", 0755),
		Download("http://get.docker.io/builds/Linux/x86_64/docker-{{ .Version }}", "{{ .InstallDir }}/bin/docker", "root", 0755),
	)
	p.AddCommands("upstart", WriteFile("/etc/init/docker.conf", dockerUpstart, "root", 0644))
}

const dockerUpstart = `exec {{ .InstallDir }}/bin/docker -d -H tcp://127.0.0.1:5432 -H unix:///var/run/docker.sock 2>&1 | logger -i -t docker
`

func (tpl *Docker) InstallDir() string {
	if tpl.Version == "" {
		panic("Version must be set")
	}
	return "/opt/docker-" + tpl.Version
}
