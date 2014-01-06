package redis

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
	Version string `urknall:"default=2.8.3"`
}

func (p *Package) InstallPath() string {
	return "/opt/redis-" + p.Version
}

func (p *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.InstallPackages("build-essential"),
		cmd.Mkdir("/opt/src/", "root", 0755),
		cmd.DownloadAndExtract(p.url(), "/opt/src/"),
		cmd.And(
			"cd /opt/src/redis-{{ .Version }}",
			"make",
			"PREFIX={{ .InstallPath }} make install",
		),
		cmd.Mkdir("/data/redis", "root", 0755),
		&Config{},
		&Upstart{RedisDir: p.InstallPath()},
	)
}

func (p *Package) WriteConfig(config string) cmd.Command {
	if e := urknall.InitializePackage(p); e != nil {
		panic(e.Error())
	}
	return cmd.WriteFile("/etc/redis.conf", config, "root", 0644)
}

func (p *Package) url() string {
	return "http://download.redis.io/releases/redis-{{ .Version }}.tar.gz"
}
