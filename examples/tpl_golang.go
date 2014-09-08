package main

import (
	"github.com/dynport/urknall"
)

type Golang struct {
	Version string `urknall:"required=true"`
}

func (golang *Golang) Render(pkg urknall.Package) {
	pkg.AddCommands("packages", InstallPackages("build-essential", "curl", "bzr", "mercurial", "git-core"))
	pkg.AddCommands("mkdir", Mkdir("{{ .InstallDir }}", "root", 0755))
	pkg.AddCommands("download",
		DownloadAndExtract("https://storage.googleapis.com/golang/go{{ .Version }}.linux-amd64.tar.gz", "{{ .InstallDir }}"),
	)
}

func (golang *Golang) InstallDir() string {
	if golang.Version == "" {
		panic("Version must bese")
	}
	return "/opt/go-" + golang.Version
}
