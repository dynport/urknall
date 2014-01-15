package apt

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type Proxy struct {
	Address string `urknall:"required=true"`
}

func (proxy *Proxy) Package(r *urknall.Runlist) {
	r.Add(
		cmd.WriteFile("/etc/apt/apt.conf.d/01proxy", `Acquire::http { Proxy "http://{{ .Address }}"; };`, "root", 0644),
		"apt-get update",
	)
}
