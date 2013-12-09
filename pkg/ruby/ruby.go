// The RubyPackage is used to provision ruby on a host.
//
// Ruby will be downloaded, extracted, configured, built, and installed to `/opt/ruby-{{ .Version }}`. If the `Bundle`
// flag is set, bundler will be installed.
package ruby

import (
	"fmt"
	. "github.com/dynport/zwo/cmd"
	"github.com/dynport/zwo/zwo"
	"strings"
)

type RubyPackage struct {
	Version     string `zwo="default=2.0.0-p247"`
	WithBundler bool
}

func (ruby *RubyPackage) Package(r *zwo.Runlist) {
	r.Add(
		InstallPackages("curl", "build-essential", "git-core",
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

func (ruby *RubyPackage) downloadURL() string {
	majorVersion := strings.Join(strings.Split(ruby.Version, ".")[0:2], ".")
	return fmt.Sprintf("http://ftp.ruby-lang.org/pub/ruby/%s/ruby-%s.tar.gz", majorVersion, ruby.Version)
}

func (ruby *RubyPackage) InstallPath() string {
	return fmt.Sprintf("/opt/ruby-%s", ruby.Version)
}

func (ruby *RubyPackage) SourcePath() string {
	return fmt.Sprintf("/opt/src/ruby-%s", ruby.Version)
}
