package main

import (
	"strings"

	"github.com/dynport/urknall"
)

// Ruby compiles and installs ruby from source
//
// Ruby will be downloaded, extracted, configured, built, and installed to `/opt/ruby-{{ .Version }}`. If the `Bundle`
// flag is set, bundler will be installed.
type Ruby struct {
	Version string `urknall:"required=true"`
	Local   bool   // install to /usr/local/bin
}

func (ruby *Ruby) Render(r urknall.Package) {
	r.AddCommands("packages",
		InstallPackages(
			"curl", "build-essential", "libyaml-dev", "libxml2-dev", "libxslt1-dev", "libreadline-dev", "libssl-dev", "zlib1g-dev",
		),
	)
	r.AddCommands("download", DownloadAndExtract("{{ .Url }}", "/opt/src"))
	r.AddCommands("build",
		And(
			"cd {{ .SourcePath }}",
			"./configure --disable-install-doc --prefix={{ .InstallPath }}",
			"make",
			"make install",
		),
	)
}

func (ruby *Ruby) Url() string {
	return "http://ftp.ruby-lang.org/pub/ruby/{{ .MinorVersion }}/ruby-{{ .Version }}.tar.gz"
}

func (ruby *Ruby) MinorVersion() string {
	return strings.Join(strings.Split(ruby.Version, ".")[0:2], ".")
}

func (ruby *Ruby) InstallPath() string {
	if ruby.Local {
		return "/usr/local"
	}
	return "/opt/ruby-{{ .Version }}"
}

func (ruby *Ruby) SourcePath() string {
	return "/opt/src/ruby-{{ .Version }}"
}
