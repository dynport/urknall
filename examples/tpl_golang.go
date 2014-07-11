package main

import (
	"github.com/dynport/urknall"
)

type Golang struct {
	Version string `urknall:"required=true"`
}

func (pkg *Golang) Render(r urknall.Package) {
	r.AddCommands("packages", InstallPackages("build-essential", "curl", "bzr", "mercurial", "git-core"))
	r.AddCommands("mkdir", Mkdir("{{ .InstallDir }}", "root", 0755))
	r.AddCommands("download",
		DownloadAndExtract("https://storage.googleapis.com/golang/go{{ .Version }}.linux-amd64.tar.gz", "{{ .InstallDir }}"),
	)
}

func (tpl *Golang) InstallDir() string {
	if tpl.Version == "" {
		panic("Version must bese")
	}
	return "/opt/go-" + tpl.Version
}
