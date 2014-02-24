package nodejs

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type Package struct {
	Version string `urknall:"default=v0.11.11"`
}

func (pkg *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.Mkdir("/opt/src", "root", 0755),
		cmd.And(
			"cd /opt/src",
			"git clone git://github.com/ry/node.git",
			"cd node",
			"git checkout {{ .Version }}",
			"./configure",
			"make -j8",
			"make install"))
}
