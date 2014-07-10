package main

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/utils"
)

type Postgres struct {
	Version string `urknall:"default=9.3.3"`
	DataDir string `urknall:"default=/data/postgres"`
}

func (p *Postgres) User() string {
	return "postgres"
}

func (pkg *Postgres) Render(r urknall.Package) {
	r.AddCommands("base",
		InstallPackages("build-essential", "openssl", "libssl-dev", "flex", "zlib1g-dev", "libxslt1-dev", "libxml2-dev", "python-dev", "libreadline-dev", "bison"),
		Mkdir("/opt/src/", "root", 0755),
		DownloadAndExtract(pkg.url(), "/opt/src/"),
		And(
			"cd /opt/src/postgresql-{{ .Version }}",
			"./configure --prefix="+pkg.InstallDir(),
			"make",
			"make install",
		),
		AddUser(pkg.User(), true),
	)
}

func (p *Postgres) InstallDir() string {
	return "/opt/postgresql-" + p.Version
}

func (pkg *Postgres) InitDbCommand() cmd.Command {
	return And(
		Mkdir(pkg.DataDir, pkg.User(), 0755),
		"su -l "+pkg.User()+" -c '"+pkg.InstallDir()+"/bin/initdb -D "+pkg.DataDir+" -E utf8 --auth-local=trust'",
	)
}

type PostgresDatabase struct {
	Name  string
	Owner string
}

func (db *PostgresDatabase) CreateCommand() string {
	cmd := "CREATE DATABASE " + db.Name
	if db.Owner != "" {
		cmd += " OWNER=" + db.Owner
	}
	return cmd
}

type PostgresUser struct {
	Name     string
	Password string
}

func (user *PostgresUser) CreateCommand() string {
	cmd := "CREATE USER " + user.Name
	if user.Password != "" {
		cmd += " WITH PASSWORD '" + user.Password + "'"
	}
	return cmd
}

func (pkg *Postgres) CreateDatabaseCommand(db *PostgresDatabase) string {
	return pkg.InstallDir() + "/bin/" + `psql -U postgres -c "` + db.CreateCommand() + `"`
}

func (pkg *Postgres) CreateUserCommand(user *PostgresUser) string {
	return pkg.InstallDir() + "/bin/" + `psql -U postgres -c "` + user.CreateCommand() + `"`
}

func (pkg *Postgres) UpstartExecCommand() cmd.Command {
	return WriteFile("/etc/init/postgres.conf", utils.MustRenderTemplate(postgresUpstart, pkg), "root", 0644)
}

const postgresUpstart = `
start on runlevel [2345]
stop on runlevel [!2345]
setuid {{ .User }}
exec {{ .InstallDir }}/bin/postgres -D {{ .DataDir }}
`

func (pkg *Postgres) url() string {
	return "http://ftp.postgresql.org/pub/source/v{{ .Version }}/postgresql-{{ .Version }}.tar.gz"
}

func (pkg *Postgres) installDir() string {
	return utils.MustRenderTemplate("/opt/postgresql-{{ .Version }}", pkg)
}

type PostGis struct {
	Version            string `urknall:"default=2.1.1"`
	PostgresInstallDir string `urknall:"required=true"`
}

func (g *PostGis) url() string {
	return "http://download.osgeo.org/postgis/source/postgis-{{ .Version }}.tar.gz"
}

func (g *PostGis) Render(r urknall.Package) {
	r.AddCommands("base",
		Mkdir("/opt/src", "root", 0755),
		DownloadAndExtract(g.url(), "/opt/src/"),
		InstallPackages("imagemagick", "libgeos-dev", "libproj-dev", "libgdal-dev"),
		Shell("cd /opt/src/postgis-{{ .Version }} && ./configure --with-pgconfig={{ .PostgresInstallDir }}/bin/pg_config --prefix=/opt/postgis-{{ .Version }} && make && make install"),
	)
}
