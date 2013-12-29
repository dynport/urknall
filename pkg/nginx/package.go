package nginx

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/utils"
)

func New(version string) *Package {
	return &Package{Version: version}
}

type Package struct {
	Version            string `urknall:"default=1.4.4"`
	HeadersMoreVersion string `urknall:"default=0.24"`
	SyslogPatchVersion string `urknall:"default=1.3.14"`
}

func (pkg *Package) Package(r *urknall.Runlist) {
	//srcDir := "/opt/src/nginx-" + pkg.Version
	syslogPatchPath := "/tmp/nginx_syslog_patch"
	fileName := "syslog_{{ .SyslogPatchVersion }}.patch"
	r.Add(
		cmd.InstallPackages("build-essential", "curl", "libpcre3", "libpcre3-dev", "libssl-dev", "libpcrecpp0", "zlib1g-dev", "libgd2-xpm-dev"),
		cmd.DownloadAndExtract(pkg.url(), "/opt/src"),
		cmd.Mkdir(syslogPatchPath, "root", 0755),
		cmd.DownloadToFile("https://raw.github.com/yaoweibin/nginx_syslog_patch/master/config", syslogPatchPath+"/config", "root", 0644),
		cmd.DownloadToFile("https://raw.github.com/yaoweibin/nginx_syslog_patch/master/"+fileName, syslogPatchPath+"/"+fileName, "root", 0644),
		cmd.And(
			"cd /opt/src/nginx-{{ .Version }}",
			"patch -p1 < "+syslogPatchPath+"/"+fileName,
		),
		cmd.DownloadToFile("https://github.com/agentzh/headers-more-nginx-module/archive/v{{ .HeadersMoreVersion }}.tar.gz", "/opt/src/headers-more-nginx-module-{{ .HeadersMoreVersion }}.tar.gz", "root", 0644),
		cmd.And(
			"cd /opt/src",
			"tar xvfz headers-more-nginx-module-{{ .HeadersMoreVersion }}.tar.gz",
		),
		cmd.And(
			"cd /opt/src/nginx-{{ .Version }}",
			"./configure --with-http_ssl_module --with-http_gzip_static_module --with-http_stub_status_module --with-http_spdy_module --add-module=/tmp/nginx_syslog_patch --add-module=/opt/src/headers-more-nginx-module-{{ .HeadersMoreVersion }} --prefix=/opt/nginx-{{ .Version }}",
			"make",
			"make install",
		),
		cmd.WriteFile("/etc/init/nginx.conf", utils.MustRenderTemplate(upstartScript, pkg), "root", 0644),
	)
}

func (pkg *Package) WriteConfigCommand(b []byte) cmd.Command {
	return cmd.WriteFile(pkg.InstallPath()+"/conf/nginx.conf", string(b), "root", 0644)
}

func (pkg *Package) BinPath() string {
	return pkg.InstallPath() + "/sbin/nginx"
}

func (pkg *Package) ReloadCommand() string {
	return utils.MustRenderTemplate("{{ . }} -t && {{ . }} -s reload", pkg.BinPath())
}

const upstartScript = `# nginx
 
description "nginx http daemon"
author "George Shammas <georgyo@gmail.com>"
 
start on (filesystem and net-device-up IFACE=lo)
stop on runlevel [!2345]
 
env DAEMON={{ .InstallPath }}/sbin/nginx
env PID=/var/run/nginx.pid
 
expect fork
respawn
respawn limit 10 5
#oom never
 
pre-start script
        $DAEMON -t
        if [ $? -ne 0 ]
                then exit $?
        fi
end script
 
exec $DAEMON
`

func (pkg *Package) url() string {
	return "http://nginx.org/download/" + pkg.fileName()
}

func (pkg *Package) InstallPath() string {
	return "/opt/nginx-" + pkg.Version
}

func (pkg *Package) fileName() string {
	return pkg.name() + ".tar.gz"
}

func (pkg *Package) name() string {
	return "nginx-" + pkg.Version
}
