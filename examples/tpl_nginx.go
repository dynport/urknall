package main

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/utils"
)

type Nginx struct {
	Version            string `urknall:"required=true"` // e.g. 1.4.7
	HeadersMoreVersion string `urknall:"default=0.24"`
	SyslogPatchVersion string `urknall:"default=1.3.14"`
	Local              bool   // install to /usr/local/nginx
	Autostart          bool
}

func (ngx *Nginx) Render(pkg urknall.Package) {
	syslogPatchPath := "/tmp/nginx_syslog_patch"
	fileName := "syslog_{{ .SyslogPatchVersion }}.patch"
	pkg.AddCommands("packages",
		InstallPackages("build-essential", "curl", "libpcre3", "libpcre3-dev", "libssl-dev", "libpcrecpp0", "zlib1g-dev", "libgd2-xpm-dev"),
	)
	pkg.AddCommands("download",
		DownloadAndExtract("{{ .Url }}", "/opt/src/"),
	)
	pkg.AddCommands("syslog_patch",
		Mkdir(syslogPatchPath, "root", 0755),
		Download("https://raw.github.com/yaoweibin/nginx_syslog_patch/master/config", syslogPatchPath+"/config", "root", 0644),
		Download("https://raw.github.com/yaoweibin/nginx_syslog_patch/master/"+fileName, syslogPatchPath+"/"+fileName, "root", 0644),
		And(
			"cd /opt/src/nginx-{{ .Version }}",
			"patch -p1 < "+syslogPatchPath+"/"+fileName,
		),
	)
	pkg.AddCommands("more_clear_headers",
		DownloadAndExtract("https://github.com/agentzh/headers-more-nginx-module/archive/v{{ .HeadersMoreVersion }}.tar.gz", "/opt/src/"),
	)
	pkg.AddCommands("build",
		And(
			"cd /opt/src/nginx-{{ .Version }}",
			"./configure --with-http_ssl_module --with-http_gzip_static_module --with-http_stub_status_module --with-http_spdy_module --add-module=/tmp/nginx_syslog_patch --add-module=/opt/src/headers-more-nginx-module-{{ .HeadersMoreVersion }} --prefix={{ .InstallDir }}",
			"make",
			"make install",
		),
	)
	pkg.AddCommands("upstart",
		WriteFile("/etc/init/nginx.conf", utils.MustRenderTemplate(nginxUpstartScript, ngx), "root", 0644),
	)
}

func (ngx *Nginx) ConfDir() string {
	return ngx.InstallDir() + "/conf"
}

func (ngx *Nginx) InstallDir() string {
	if ngx.Local {
		return "/usr/local/nginx"
	}
	if ngx.Version == "" {
		panic("Version must be set")
	}
	return "/opt/nginx-" + ngx.Version
}

func (ngx *Nginx) BinPath() string {
	return ngx.InstallDir() + "/sbin/nginx"
}

func (ngx *Nginx) ReloadCommand() string {
	return utils.MustRenderTemplate("{{ . }} -t && {{ . }} -s reload", ngx.BinPath())
}

const nginxUpstartScript = `# nginx
 
description "nginx http daemon"
author "George Shammas <georgyo@gmail.com>"
 
{{ if .Autostart }}
start on (filesystem and net-device-up IFACE=lo)
stop on runlevel [!2345]
{{ end }}
 
env DAEMON={{ .InstallDir }}/sbin/nginx
env PID=/var/run/nginx.pid
 
respawn
respawn limit 10 5
#oom never
 
pre-start script
        $DAEMON -t
        if [ $? -ne 0 ]
                then exit $?
        fi
end script
 
exec $DAEMON -g "daemon off;"
`

func (ngx *Nginx) Url() string {
	return "http://nginx.org/download/nginx-{{ .Version }}.tar.gz"
}
