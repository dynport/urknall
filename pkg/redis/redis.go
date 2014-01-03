package redis

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

func New(version string) *Package {
	return &Package{
		Version: version,
	}
}

type Package struct {
	Version string `urknall:"default=2.8.3"`
}

func (p *Package) InstallPath() string {
	return "/opt/redis-{{ .Version }}"
}

func (p *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.InstallPackages("build-essential"),
		cmd.Mkdir("/opt/src/", "root", 0755),
		cmd.DownloadAndExtract(p.url(), "/opt/src/"),
		cmd.And(
			"cd /opt/src/redis-{{ .Version }}",
			"make",
			"PREFIX={{ .InstallPath }} make install",
		),
		cmd.Mkdir("/data/redis", "root", 0755),
		cmd.WriteFile("/etc/redis.conf", cfg, "root", 0644),
		cmd.WriteFile("/etc/init/redis.conf", upstart, "root", 0644),
	)
}

func (p *Package) url() string {
	return "http://download.redis.io/releases/redis-{{ .Version }}.tar.gz"
}

const upstart = `
pre-start script
	sysctl vm.overcommit_memory=1
end script
exec {{ .InstallPath }}/bin/redis-server /etc/redis.conf
respawn
respawn limit 10 60
`

const cfg = `daemonize no
port 6379
timeout 0
tcp-keepalive 0
loglevel notice
syslog-enabled yes
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
