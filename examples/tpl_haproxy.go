package main

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall"
)

type HAProxy struct {
	Version string `urknall:"required=true"`
}

func (p *HAProxy) url() string {
	return "http://haproxy.1wt.eu/download/{{ .MinorVersion }}/src/haproxy-{{ .Version }}.tar.gz"
}

func (p *HAProxy) MinorVersion() string {
	parts := strings.Split(p.Version, ".")
	if len(parts) == 3 {
		return strings.Join(parts[0:2], ".")
	}
	panic(fmt.Sprintf("unable to extract minor version from %q", p.Version))
}

func (p *HAProxy) InstallDir() string {
	if p.Version == "" {
		panic("Version must be set")
	}
	return "/opt/haproxy-" + p.Version
}

func (p *HAProxy) Render(r urknall.Package) {
	r.AddCommands("base",
		InstallPackages("curl", "build-essential", "libpcre3-dev"),
		Mkdir("/opt/src/", "root", 0755),
		DownloadAndExtract(p.url(), "/opt/src/"),
		Mkdir("{{ .InstallDir }}/sbin", "root", 0755),
		Shell("cd /opt/src/haproxy-{{ .Version }} && make TARGET=linux25 USER_STATIC_PCRE=1 && cp ./haproxy {{ .InstallDir }}/sbin/"),
		WriteFile("/etc/init/haproxy.conf", initScript, "root", 0755),
	)
}

const initScript = `description "Properly handle haproxy"

start on (filesystem and net-device-up IFACE=lo)

env PID_PATH=/var/run/haproxy.pid
env BIN_PATH={{ .InstallDir }}/sbin/haproxy

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
