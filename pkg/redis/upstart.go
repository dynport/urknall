package redis

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type Upstart struct {
	Name        string `urknall:"default=redis"`
	RedisConfig string `urknall:"default=/etc/redis.conf"`
	RedisDir    string `urknall:"required=true"`
}

func (u *Upstart) Package(r *urknall.Runlist) {
	r.Add(
		cmd.WriteFile("/etc/init/{{ .Name }}.conf", upstart, "root", 0644),
	)
	return
}

const upstart = `
pre-start script
	sysctl vm.overcommit_memory=1
end script
exec {{ .RedisDir }}/bin/redis-server {{ .RedisConfig }}
respawn
respawn limit 10 60
`
