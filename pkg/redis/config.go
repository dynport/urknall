package redis

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type Config struct {
	Port        int    `urknall:"default=6379"`
	Path        string `urknall:"default=/etc/redis.conf"`
	SyslogIdent string `urknall:"default=redis"`
}

func (c *Config) Package(r *urknall.Runlist) {
	r.Add(
		cmd.WriteFile(c.Path, cfg, "root", 0644),
	)
}

const cfg = `daemonize no
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
