// The Ruby package is used to provision ruby on a host.
//
// Ruby will be downloaded, extracted, configured, built, and installed to `/opt/ruby-{{ .Version }}`. If the `Bundle`
// flag is set, bundler will be installed.
package main

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall"
)

func NewRuby(version string) *Ruby {
	return &Ruby{Version: version}
}

type Ruby struct {
	Version     string `urknall:"default=2.0.0-p247"`
	WithBundler bool
	Local       bool // install to /usr/local/bin
}

func (ruby *Ruby) PkgVersion() string {
	return ruby.Version
}

func (ruby *Ruby) Name() string {
	return "ruby"
}

func (ruby *Ruby) PackageDependencies() []string {
	return []string{"libyaml-0-2", "libxml2", "libxslt1.1", "libreadline6", "libssl1.0.0", "zlib1g"}
}

func (ruby *Ruby) Package(r *urknall.Package) {
	r.Add(
		InstallPackages("curl", "build-essential",
			"libyaml-dev", "libxml2-dev", "libxslt1-dev",
			"libreadline-dev", "libssl-dev", "zlib1g-dev"))

	r.Add(
		DownloadAndExtract(ruby.downloadURL(), "/opt/src"))

	r.Add(
		And("cd {{ .SourcePath }}",
			"./configure --disable-install-doc --prefix={{ .InstallPath }}",
			"make",
			"make install"))

	if ruby.WithBundler {
		r.Add("{{ .InstallPath }}/bin/gem install bundler")
	}
}

func (ruby *Ruby) downloadURL() string {
	majorVersion := strings.Join(strings.Split(ruby.Version, ".")[0:2], ".")
	return fmt.Sprintf("http://ftp.ruby-lang.org/pub/ruby/%s/ruby-%s.tar.gz", majorVersion, ruby.Version)
}

func (ruby *Ruby) InstallPath() string {
	if ruby.Local {
		return "/usr/local"
	}
	return fmt.Sprintf("/opt/ruby-%s", ruby.Version)
}

func (ruby *Ruby) SourcePath() string {
	return fmt.Sprintf("/opt/src/ruby-%s", ruby.Version)
}
