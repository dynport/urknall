package main

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

func NewRedis(version string) *Redis {
	return &Redis{Version: version}
}

type Redis struct {
	Version   string `urknall:"default=2.8.3"`
	Autostart bool
}

func (p *Redis) InstallPath() string {
	return "/opt/redis-" + p.Version
}

func (p *Redis) Package(r *urknall.Package) {
	r.Add(
		InstallPackages("build-essential"),
		Mkdir("/opt/src/", "root", 0755),
		DownloadAndExtract(p.url(), "/opt/src/"),
		And(
			"cd /opt/src/redis-{{ .Version }}",
			"make",
			"PREFIX={{ .InstallPath }} make install",
		),
		Mkdir("/data/redis", "root", 0755),
		&RedisConfig{},
		&RedisUpstart{RedisDir: p.InstallPath(), Autostart: p.Autostart},
	)
}

func (p *Redis) WriteConfig(config string) cmd.Command {
	return WriteFile("/etc/redis.conf", config, "root", 0644)
}

func (p *Redis) url() string {
	return "http://download.redis.io/releases/redis-{{ .Version }}.tar.gz"
}

type RedisConfig struct {
	Port        int    `urknall:"default=6379"`
	Path        string `urknall:"default=/etc/redis.conf"`
	SyslogIdent string `urknall:"default=redis"`
}

func (c *RedisConfig) Package(r *urknall.Package) {
	r.Add(
		WriteFile(c.Path, redisCfg, "root", 0644),
	)
}

const redisCfg = `daemonize no
port {{ .Port }}
timeout 0
tcp-keepalive 0
loglevel notice
syslog-enabled yes
syslog-ident {{ .SyslogIdent }}
databases 16
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /data/redis
slave-serve-stale-data yes
slave-read-only yes
repl-disable-tcp-nodelay no
slave-priority 100
appendonly yes
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
lua-time-limit 5000
slowlog-log-slower-than 10000
slowlog-max-len 128
notify-keyspace-events ""
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-entries 512
list-max-ziplist-value 64
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
activerehashing yes
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit slave 256mb 64mb 60
client-output-buffer-limit pubsub 32mb 8mb 60
hz 10
aof-rewrite-incremental-fsync yes
`

type RedisUpstart struct {
	Name        string `urknall:"default=redis"`
	RedisConfig string `urknall:"default=/etc/redis.conf"`
	RedisDir    string `urknall:"required=true"`
	Autostart   bool
}

func (u *RedisUpstart) Package(r *urknall.Package) {
	r.Add(
		WriteFile("/etc/init/{{ .Name }}.conf", redisUpstart, "root", 0644),
	)
	return
}

const redisUpstart = `
{{ if .Autostart }}
start on (local-filesystems and net-device-up IFACE!=lo)
{{ end }}
pre-start script
	sysctl vm.overcommit_memory=1
end script
exec {{ .RedisDir }}/bin/redis-server {{ .RedisConfig }}
respawn
respawn limit 10 60
`
