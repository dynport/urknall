package main

import "github.com/dynport/urknall"

type PostGis struct {
	Version            string `urknall:"required=true"`
	PostgresInstallDir string `urknall:"required=true"`
}

func (g *PostGis) url() string {
	return "http://download.osgeo.org/postgis/source/postgis-{{ .Version }}.tar.gz"
}

func (g *PostGis) Render(r urknall.Package) {
	r.AddCommands("packages",
		InstallPackages("imagemagick", "libgeos-dev", "libproj-dev", "libgdal-dev"),
	)
	r.AddCommands("download",
		DownloadAndExtract(g.url(), "/opt/src/"),
	)
	r.AddCommands("build",
		And(
			"cd /opt/src/postgis-{{ .Version }}",
			"./configure --with-pgconfig={{ .PostgresInstallDir }}/bin/pg_config --prefix={{ .InstallDir }}",
			"make",
			"make install",
		),
	)
}

func (tpl *PostGis) InstallDir() string {
	if tpl.Version == "" {
		panic("Version must be set")
	}
	return "/opt/postgis-" + tpl.Version
}
