package main

import (
	"github.com/dynport/urknall"
)

type Golang struct {
	Version string `urknall:"default=1.2"`
}

func NewGolang(version string) *Golang {
	return &Golang{Version: version}
}

func (pkg *Golang) Render(r urknall.Package) {
	url := "http://go.googlecode.com/files/go{{ .Version }}.linux-amd64.tar.gz"
	r.AddCommands("base",
		InstallPackages("build-essential", "curl", "bzr", "mercurial", "git-core"),
		Mkdir("/opt/go-{{ .Version }}", "root", 0755),
		DownloadAndExtract(url, "/opt/go-{{ .Version }}/"),
	)
}

func (pkg *Golang) Goroot() string {
	return "/opt/go-" + pkg.Version + "/go"
}
