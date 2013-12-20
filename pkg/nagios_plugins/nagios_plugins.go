package nagios_plugins

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/utils"
)

type Package struct {
	Version string `urknall:"default=1.5"`
}

func (plugins *Package) InstallPath() string {
	if plugins.Version == "" {
		panic(".Version must be set for NagiosPlugins")
	}
	return utils.MustRenderTemplate("/opt/nagios-plugins-{{ .Version }}", plugins)
}

func (plugins *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.InstallPackages("libssl-dev", "openssl", "file"),
		cmd.DownloadAndExtract(plugins.url(), "/opt/src"),
		cmd.And(
			"cd /opt/src/nagios-plugins-{{ .Version }}",
			"./configure --prefix={{ .InstallPath }} --with-ssl-lib=/usr/lib/x86_64-linux-gnu --with-ssl-lib=/usr/lib",
			"make",
			"make install",
		),
	)
}

func (plugins *Package) url() string {
	return "https://www.nagios-plugins.org/download/nagios-plugins-{{ .Version }}.tar.gz"
}
