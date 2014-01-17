package haproxy

import (
	"fmt"
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"strings"
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
	return "http://haproxy.1wt.eu/download/1.4/src/haproxy-1.4.24.tar.gz"
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
		cmd.WriteFile("/etc/init.d/haproxy", initScript, "root", 0755),
	)
}

const initScript = `#!/usr/bin/env bash

### BEGIN INIT INFO
# Provides:          Urknall provided this script to provide a service.
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Start daemon at boot time
# Description:       Enable service provided by daemon.
### END INIT INFO

PID_PATH=/var/run/haproxy.pid
BIN_PATH={{ .InstallPath }}/sbin/haproxy

case $1 in
  "status")
    start-stop-daemon -p $PID_PATH --status
    code=$?
    case $code in
      0)
        echo "STATUS: running"
        ;;
      1)
        echo "STATUS: NOT running (but pid exists)"
        ;;
      3)
        echo "STATUS: NOT running"
        ;;
      4)
        echo "STATUS: UNKNOWN"
        ;;
    esac
    exit $code
    ;;
  "stop")
    start-stop-daemon -p $PID_PATH --stop
    ;;
  "start")
    $BIN_PATH -f /etc/haproxy.cfg -D -p $PID_PATH
    ;;

  "configtest")
    $BIN_PATH -f /etc/haproxy.cfg -D -p $PID_PATH -c
    ;;

  "reload")
    $BIN_PATH -f /etc/haproxy.cfg -D -p $PID_PATH -sf $(cat $PID_PATH)
    ;;

  *)
    echo "ERROR: command $1 unknown. Support commands: status, start, stop"
    exit 5
    ;;
esac
`
