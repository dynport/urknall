package main

import "github.com/dynport/urknall"

type PostGis struct {
	Version            string `urknall:"required=true"`
	PostgresInstallDir string `urknall:"required=true"`
}

func (pgis *PostGis) url() string {
	return "http://download.osgeo.org/postgis/source/postgis-{{ .Version }}.tar.gz"
}

func (pgis *PostGis) Render(pkg urknall.Package) {
	pkg.AddCommands("packages",
		InstallPackages("imagemagick", "libgeos-dev", "libproj-dev", "libgdal-dev"),
	)
	pkg.AddCommands("download",
		DownloadAndExtract(pgis.url(), "/opt/src/"),
	)
	pkg.AddCommands("build",
		And(
			"cd /opt/src/postgis-{{ .Version }}",
			"./configure --with-pgconfig={{ .PostgresInstallDir }}/bin/pg_config --prefix={{ .InstallDir }}",
			"make",
			"make install",
		),
	)
}

func (pgis *PostGis) InstallDir() string {
	if pgis.Version == "" {
		panic("Version must be set")
	}
	return "/opt/postgis-" + pgis.Version
}
