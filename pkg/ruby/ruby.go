// The Ruby package is used to provision ruby on a host.
//
// Ruby will be downloaded, extracted, configured, built, and installed to `/opt/ruby-{{ .Version }}`. If the `Bundle`
// flag is set, bundler will be installed.
package ruby

import (
	"fmt"
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"strings"
)

func New(version string) *Package {
	return &Package{
		Version: version,
	}
}

type Package struct {
	Version     string `urknall:"default=2.0.0-p247"`
	WithBundler bool
}

func (ruby *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.InstallPackages("curl", "build-essential",
			"libyaml-dev", "libxml2-dev", "libxslt1-dev",
			"libreadline-dev", "libssl-dev", "zlib1g-dev"))

	r.Add(
		cmd.DownloadAndExtract(ruby.downloadURL(), "/opt/src"))

	r.Add(
		cmd.And("cd {{ .SourcePath }}",
			"./configure --disable-install-doc --prefix={{ .InstallPath }}",
			"make",
			"make install"))

	if ruby.WithBundler {
		r.Add("{{ .InstallPath }}/bin/gem install bundler")
	}
}

func (ruby *Package) downloadURL() string {
	majorVersion := strings.Join(strings.Split(ruby.Version, ".")[0:2], ".")
	return fmt.Sprintf("http://ftp.ruby-lang.org/pub/ruby/%s/ruby-%s.tar.gz", majorVersion, ruby.Version)
}

func (ruby *Package) InstallPath() string {
	return fmt.Sprintf("/opt/ruby-%s", ruby.Version)
}

func (ruby *Package) SourcePath() string {
	return fmt.Sprintf("/opt/src/ruby-%s", ruby.Version)
}
