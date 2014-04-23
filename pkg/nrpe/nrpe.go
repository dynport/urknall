package nrpe

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
)

type Command struct {
	Name    string
	Command string
}

type Package struct {
	Version      string `urknall:"default=2.15"`
	Commands     []*Command
	AllowedHosts string
}

func (nrpe *Package) Package(r *urknall.Package) {
	r.Add(
		cmd.AddUser("nagios", true),
		cmd.Mkdir("/var/run/nagios", "nagios", 0755),
		cmd.DownloadAndExtract(nrpe.url(), "/opt/src"),
		cmd.And(
			"cd /opt/src/nrpe-{{ .Version }}",
			"./configure --with-ssl=/usr/bin/openssl --with-ssl-lib=/usr/lib/x86_64-linux-gnu --enable-command-args --prefix=/opt/nrpe-{{ .Version }}",
			"make",
			"make all",
			"make install",
			"make install-plugin install-daemon install-daemon-config",
		),
		cmd.WriteFile("/etc/init/nrpe.conf", upstartScript, "root", 0644),
		cmd.WriteFile("/opt/nrpe-{{ .Version }}/etc/nrpe.cfg", nrpeConfig, "nagios", 0644),
		"{ service nrpe status && service nrpe restart; } || service nrpe start",
	)
}

const upstartScript = `start on filesystem
stop on runlevel [06]

respawn

expect fork
exec /opt/nrpe-{{ .Version }}/bin/nrpe -c /opt/nrpe-{{ .Version }}/etc/nrpe.cfg -d
`

const nrpeConfig = `log_facility=daemon
pid_file=/var/run/nagios/nrpe.pid
server_port=5666
nrpe_user=nagios
nrpe_group=nagios
{{ with .AllowedHosts }}
allowed_hosts={{ . }}
{{ end }}
dont_blame_nrpe=1
debug=0
command_timeout=60
connection_timeout=300

{{ range .Commands }}
command[{{ .Name }}]={{ .Command }}
{{ end }}
`

func (nrpe *Package) url() string {
	return "http://prdownloads.sourceforge.net/sourceforge/nagios/nrpe-{{ .Version }}.tar.gz"
}
