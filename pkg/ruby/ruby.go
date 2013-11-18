// The RubyPackage is used to provision ruby on a host.
//
// TODO: add prefix support
package ruby

import (
	"fmt"
	. "github.com/dynport/zwo/cmd"
	"github.com/dynport/zwo/zwo"
	"strings"
)

type RubyPackage struct {
	Version string `json:"version" default:"2.0.0-p247" desc:"package version"`
}

func (ruby *RubyPackage) Compile(r *zwo.Runlist) {
	r.Execute(
		InstallPackages("curl", "build-essential", "git-core",
			"libyaml-dev", "libxml2-dev", "libxslt1-dev",
			"libreadline-dev", "libssl-dev", "zlib1g-dev"))
	r.Execute(
		And("mkdir -p /opt/downloads",
			"cd /opt/downloads",
			"curl -SsfLO \"{{ .getDownloadURL }}\""))

	r.Execute(
		And("mkdir -p /opt/src",
			"cd /opt/src",
			"tar xvfz {{ .getSourceFilename }}"))

	r.Execute(
		And("cd {{ .getSoucePath }}",
			"./configure --disable-install-doc",
			"make"))

	r.Execute(
		And(
			"cd {{ .getSoucePath }}",
			"make install",
			"ln -nfs /opt/{{ .getPathSegment }} /opt/ruby"))

	r.AddFile("/root/.profile.d/ruby", "export PATH=/opt/ruby/bin:$PATH", "", 0)
}

func (ruby *RubyPackage) getSourceFilename() string {
	return fmt.Sprintf("ruby-%s.tar.gz", ruby.Version)
}

func (ruby *RubyPackage) getDownloadURL() string {
	majorVersion := strings.Join(strings.Split(ruby.Version, ".")[0:2], ".")
	return fmt.Sprintf("ftp://ftp.ruby-lang.org/pub/ruby/%s/%s", majorVersion, ruby.getSourceFilename())
}

func (ruby *RubyPackage) getPathSegment() string {
	return fmt.Sprintf("ruby-%s", ruby.Version)
}

func (ruby *RubyPackage) getSourcePath() string {
	return fmt.Sprintf("/opt/src/%s", ruby.getPathSegment())
}

func cdShellCmd(path string) string {
	return fmt.Sprintf("cd %s", path)
}
