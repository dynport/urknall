package main

import "github.com/dynport/urknall"

type ElasticSearch struct {
	Version     string `urknall:"required=true"`
	ClusterName string `urknall:"default=elasticsearch"`
	DataPath    string `urknall:"default=/data/elasticsearch"`

	// optional
	SyslogHost     string
	DiscoveryHosts string
	LogPath        string
	NodeName       string
}

func (es *ElasticSearch) Render(pkg urknall.Package) {
	pkg.AddCommands("java7",
		InstallPackages("openjdk-7-jdk"),
		Shell("update-alternatives --set java /usr/lib/jvm/java-7-openjdk-amd64/jre/bin/java"),
	)
	pkg.AddCommands("download", DownloadAndExtract("{{ .Url }}", "/opt/"))
	pkg.AddCommands("user", AddUser("elasticsearch", true))
	pkg.AddCommands("mkdir", Mkdir(es.DataPath, "elasticsearch", 0755))
	pkg.AddCommands("config",
		WriteFile("{{ .InstallDir }}/config/elasticsearch.yml", elasticSearchConfig, "root", 0644),
		WriteFile("{{ .InstallDir }}/config/logging.yml", elasticSearchConfigLogger, "root", 0644),
		WriteFile("/etc/init/elasticsearch.conf", elasticSearchUpstart, "root", 0644),
	)
}

func (es *ElasticSearch) Url() string {
	return "https://download.elasticsearch.org/elasticsearch/elasticsearch/elasticsearch-{{ .Version }}.tar.gz"
}

func (es *ElasticSearch) InstallDir() string {
	if es.Version == "" {
		panic("Version must be set")
	}
	return "/opt/elasticsearch-" + es.Version
}

const elasticSearchUpstart = `
{{ with .DataPath }}
pre-start script
	mkdir -p {{ . }}
end script
{{ end }}

exec {{ .InstallDir }}/bin/elasticsearch -f
`

const elasticSearchConfigLogger = `
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

const elasticSearchConfig = `
path.data: {{ .DataPath }}
path.logs: {{ .DataPath }}/logs
{{ with .NodeName }}node.name: {{ . }}{{ end }}
{{ with .ClusterName }}cluster.name: {{ . }}{{ end }}
{{ with .DiscoveryHosts }}discovery.zen.ping.unicast.hosts: {{ . }}{{ end }}
`
