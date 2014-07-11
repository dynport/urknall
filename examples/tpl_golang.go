package main

import (
	"github.com/dynport/urknall"
)

type Golang struct {
	Version string `urknall:"required=true"`
}

func (pkg *Golang) Render(r urknall.Package) {
	r.AddCommands("packages", InstallPackages("build-essential", "curl", "bzr", "mercurial", "git-core"))
	r.AddCommands("mkdir", Mkdir("{{ .InstallPath }}", "root", 0755))
	r.AddCommands("download",
		DownloadAndExtract("https://storage.googleapis.com/golang/go{{ .Version }}.linux-amd64.tar.gz", "{{ .InstallPath }}"),
	)
}

func (tpl *Golang) InstallPath() string {
	return "/opt/go-{{ .Version }}"
}
