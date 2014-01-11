package golang

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type Package struct {
	Version string `urknall:"default=1.2"`
}

func New(version string) *Package {
	return &Package{
		Version: version,
	}
}

func (pkg *Package) Package(r *urknall.Runlist) {
	url := "https://go.googlecode.com/files/go{{ .Version }}.linux-amd64.tar.gz"
	r.Add(
		cmd.InstallPackages("build-essential", "curl", "bzr", "mercurial", "git-core"),
		cmd.Mkdir("/opt/go-{{ .Version }}", "root", 0755),
		cmd.DownloadAndExtract(url, "/opt/go-{{ .Version }}/"),
	)
}

func (pkg *Package) Goroot() string {
	return "/opt/go-" + pkg.Version + "/go"
}
