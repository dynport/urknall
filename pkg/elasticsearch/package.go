package elasticsearch

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
	Version     string `urknall:"default=0.90.9"`
	ClusterName string `urknall:"default=elasticsearch"`
	DataPath    string `urknall:"default=/data/elasticsearch"`

	// optional
	SyslogHost     string
	DiscoveryHosts string
	LogPath        string
	NodeName       string
}

func (p *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.InstallPackages("openjdk-6-jdk"),
		cmd.DownloadAndExtract(p.url(), "/opt/"),
		cmd.AddUser("elasticsearch", true),
		cmd.Mkdir(p.DataPath, "elasticsearch", 0755),
		cmd.WriteFile("{{ .InstallPath }}/config/elasticsearch.yml", config, "root", 0644),
		cmd.WriteFile("{{ .InstallPath }}/config/logging.yml", configLogger, "root", 0644),
		cmd.WriteFile("/etc/init/elasticsearch.conf", upstart, "root", 0644),
	)
}

func (p *Package) url() string {
	return "https://download.elasticsearch.org/elasticsearch/elasticsearch/elasticsearch-{{ .Version }}.tar.gz"
}

func (p *Package) InstallPath() string {
	return "/opt/elasticsearch-" + p.Version
}

const upstart = `
setuid elasticsearch
exec {{ .InstallPath }}/bin/elasticsearch -f
`

const configLogger = `
rootLogger: DEBUG, syslog
logger:
  # log action execution errors for easier debugging
  action: DEBUG
  # reduce the logging for aws, too much is logged under the default INFO
  com.amazonaws: WARN

  index.search.slowlog: TRACE{{ with .SyslogHost }}, syslog{{ end }}
  index.indexing.slowlog: TRACE{{ with .SyslogHost }}, syslog{{ end }}

additivity:
  index.search.slowlog: false
  index.indexing.slowlog: false


{{ with .SyslogHost }}
appender:
  syslog:
      type: syslog
      syslogHost: {{ . }}:514
      facility: local0
      layout:
        type: pattern
        conversionPattern: "[%-5p] [%-25c] %m%n"
{{ end }}
`

const config = `
path.data: {{ .DataPath }}
path.logs: {{ .DataPath }}/logs
{{ with .NodeName }}node.name: {{ . }}{{ end }}
{{ with .ClusterName }}cluster.name: {{ . }}{{ end }}
{{ with .DiscoveryHosts }}discovery.zen.ping.unicast.hosts: {{ . }}{{ end }}
`
