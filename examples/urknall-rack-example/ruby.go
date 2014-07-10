package main

import (
	"strings"

	"github.com/dynport/urknall"
)

// Ruby is a urknall.Template to install ruby from source
//
// Version is a required variable which is used when "rendering" steps/commands
// All public attributes and methods of a Template can be used when rendering
type Ruby struct {
	Version string `urknall:"required=true"`
}

// steps taken from from https://gorails.com/setup/ubuntu/14.04 ("from source")
func (tpl *Ruby) Render(p urknall.Package) {
	// install packages required by ruby/rails
	p.AddCommands("packages", InstallPackages(
		"git-core", "curl", "zlib1g-dev", "build-essential", "libssl-dev",
		"libreadline-dev", "libyaml-dev", "libsqlite3-dev", "sqlite3",
		"libxml2-dev", "libxslt1-dev", "libcurl4-openssl-dev", "python-software-properties",
	),
	)
	p.AddCommands("download",
		// create src directory
		Mkdir("/opt/src/", "root", 0755),

		// download ruby source to /opt/src/ with user=root and chmod=0644
		Download( //
			"http://ftp.ruby-lang.org/pub/ruby/{{ .MinorVersion }}/ruby-{{ .Version }}.tar.gz",
			"/opt/src/",
			"root", 644,
		),
	)

	// execute the build steps in one concatenated command (with &&)
	p.AddCommands("build",
		And(
			"cd /opt/src/",
			"tar xvfz ruby-{{ .Version }}.tar.gz",
			"cd ruby-{{ .Version }}",
			"./configure --disable-install-doc",
			"make -j 8",
			"make install",
		),
	)
}

func (r *Ruby) MinorVersion() string {
	parts := strings.Split(r.Version, ".")
	if len(parts) > 2 {
		return strings.Join(parts[0:2], ".")
	}
	panic("could not extract minor version from " + r.Version)
}
