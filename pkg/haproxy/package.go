package haproxy

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type Package struct {
	Version string `urknall:"default=1.4.24"`
}

func New(version string) *Package {
	return &Package{
		Version: version,
	}
}

func (p *Package) url() string {
	return "http://haproxy.1wt.eu/download/1.4/src/haproxy-" + p.Version + ".tar.gz"
}

func (p *Package) minorVersion() string {
	parts := strings.Split(p.Version, ".")
	if len(parts) == 3 {
		return strings.Join(parts[0:2], ".")
	}
	panic(fmt.Sprintf("unable to extract minor version from %q", p.Version))
}

func (p *Package) InstallPath() string {
	return "/opt/haproxy-" + p.Version
}

func (p *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.InstallPackages("curl", "build-essential", "libpcre3-dev"),
		cmd.Mkdir("/opt/src/", "root", 0755),
		cmd.DownloadAndExtract(p.url(), "/opt/src/"),
		cmd.Mkdir("{{ .InstallPath }}/sbin", "root", 0755),
		"cd /opt/src/haproxy-{{ .Version }} && make TARGET=linux25 USER_STATIC_PCRE=1 && cp ./haproxy {{ .InstallPath }}/sbin/",
		cmd.WriteFile("/etc/init/haproxy.conf", initScript, "root", 0755),
	)
}

const initScript = `description "Properly handle haproxy"

start on (filesystem and net-device-up IFACE=lo)

env PID_PATH=/var/run/haproxy.pid
env BIN_PATH={{ .InstallPath }}/sbin/haproxy

script
exec /bin/bash <<EOF
  $BIN_PATH -f /etc/haproxy.cfg -D -p $PID_PATH

  trap "$BIN_PATH -f /etc/haproxy.cfg -p $PID_PATH -D -sf \$(cat $PID_PATH)" SIGHUP
  trap "kill -TERM \$(cat $PID_PATH) && exit 0" SIGTERM SIGINT

  while true; do # Iterate to keep job running.
    sleep 1 # Don't sleep to long as signals will not be handled during sleep.
  done
EOF
end script`
