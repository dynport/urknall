// The RubyPackage is used to provision ruby on a host.
//
// TODO: add prefix support
package ruby

import (
	"fmt"
	"github.com/dynport/zwo/zwo"
	"strings"
)

type RubyPackage struct {
	Version string `json:"version" default:"2.0.0-p247" desc:"package version"`
}

func (ruby *RubyPackage) Compile(r *zwo.Runlist) (e error) {
	if e = ruby.installPackages(r); e != nil {
		return e
	}

	if e = ruby.downloadAndExtractRubySource(r); e != nil {
		return e
	}

	if e = ruby.buildRubyFromSource(r); e != nil {
		return e
	}

	if e = ruby.installRuby(r); e != nil {
		return e
	}
	return nil
}

func (ruby *RubyPackage) installPackages(r *zwo.Runlist) (e error) {
	packages := []string{
		"curl", "build-essential", "git-core",
		"libyaml-dev", "libxml2-dev", "libxslt1-dev",
		"libreadline-dev", "libssl-dev", "zlib1g-dev",
	}
	return r.AddCommands(zwo.InstallPackages(packages...))
}

func (ruby *RubyPackage) downloadAndExtractRubySource(r *zwo.Runlist) (e error) {
	return r.AddCommands(
		zwo.And(
			zwo.Execute("mkdir -p /opt/downloads"),
			zwo.Execute("cd /opt/downloads"),
			zwo.Execute("curl -SsfLO \"{{ .getDownloadURL }}\"")),
		zwo.And(
			zwo.Execute("mkdir -p /opt/src"),
			zwo.Execute("cd /opt/src"),
			zwo.Execute("tar xvfz {{ .getSourceFilename }}")))
}

func (ruby *RubyPackage) buildRubyFromSource(r *zwo.Runlist) (e error) {
	return r.AddCommands(
		zwo.And(
			zwo.Execute("cd {{ .getSoucePath }}"),
			zwo.Execute("./configure --disable-install-doc"),
			zwo.Execute("make")))
}

func (ruby *RubyPackage) installRuby(r *zwo.Runlist) (e error) {
	e = r.AddCommands(
		zwo.And(
			zwo.Execute("cd {{ .getSoucePath }}"),
			zwo.Execute("make install"),
			zwo.Execute("ln -nfs /opt/{{ .getPathSegment }} /opt/ruby")))
	if e != nil {
		return e
	}
	return r.AddFiles(zwo.WriteFile("/root/.profile.d/ruby", "export PATH=/opt/ruby/bin:$PATH", "", 0))
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
