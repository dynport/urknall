package main

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall"
)

type HAProxy struct {
	Version string `urknall:"default=1.4.24"`
}

func NewHAProxy(version string) *HAProxy {
	return &HAProxy{Version: version}
}

func (p *HAProxy) url() string {
	return "http://haproxy.1wt.eu/download/1.4/src/haproxy-" + p.Version + ".tar.gz"
}

func (p *HAProxy) minorVersion() string {
	parts := strings.Split(p.Version, ".")
	if len(parts) == 3 {
		return strings.Join(parts[0:2], ".")
	}
	panic(fmt.Sprintf("unable to extract minor version from %q", p.Version))
}

func (p *HAProxy) InstallPath() string {
	return "/opt/haproxy-" + p.Version
}

func (p *HAProxy) Package(r urknall.Package) {
	r.AddCommands("base",
		InstallPackages("curl", "build-essential", "libpcre3-dev"),
		Mkdir("/opt/src/", "root", 0755),
		DownloadAndExtract(p.url(), "/opt/src/"),
		Mkdir("{{ .InstallPath }}/sbin", "root", 0755),
		Shell("cd /opt/src/haproxy-{{ .Version }} && make TARGET=linux25 USER_STATIC_PCRE=1 && cp ./haproxy {{ .InstallPath }}/sbin/"),
		WriteFile("/etc/init/haproxy.conf", initScript, "root", 0755),
	)
}

const initScript = `description "Properly handle haproxy"

start on (filesystem and net-device-up IFACE=lo)

env PID_PATH=/var/run/haproxy.pid
env BIN_PATH={{ .InstallPath }}/sbin/haproxy

script
exec /bin/bash <<EOF

reload() {
  $BIN_PATH -f /etc/haproxy.cfg -p $PID_PATH -D -sf \$(cat $PID_PATH)
}

stop() {
  kill -TERM \$(cat $PID_PATH)
  exit 0
}

trap 'reload' SIGHUP
trap 'stop' SIGTERM SIGINT

$BIN_PATH -f /etc/haproxy.cfg -D -p $PID_PATH

while true; do # Iterate to keep job running.
  sleep 1 # Don't sleep to long as signals will not be handled during sleep.
done
EOF
end script`
