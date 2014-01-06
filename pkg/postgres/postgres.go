package postgres

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/utils"
)

const user = "postgres"

type Package struct {
	Version string `urknall:"default=9.3.2"`
}

func (pkg *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.InstallPackages("build-essential", "libaio1", "libssl-dev", "perl"),
		cmd.DownloadAndExtract(pkg.url(), "/opt"),
		cmd.And(
			"cd /opt/src/postgresql-{{ .Version }}",
			"./configure --prefix=/opt/postgresql-{{ .Version }}",
			"make",
			"make install",
		),
		cmd.AddUser(user, true),
	)
}

func (pkg *Package) InitDbCommand(dataDir string) cmd.Command {
	return cmd.And(
		cmd.Mkdir(dataDir, user, 0755),
		"su -l "+user+" -c '"+pkg.installDir()+"/bin/initdb -D "+dataDir+" -E utf8 --auth-local=trust'",
	)
}

type Database struct {
	Name  string
	Owner string
}

func (db *Database) CreateCommand() string {
	cmd := "CREATE DATABASE " + db.Name
	if db.Owner != "" {
		cmd += " OWNER=" + db.Owner
	}
	return cmd
}

type User struct {
	Name     string
	Password string
}

func (user *User) CreateCommand() string {
	cmd := "CREATE USER " + user.Name
	if user.Password != "" {
		cmd += " WITH PASSWORD '" + user.Password + "'"
	}
	return cmd
}

func (pkg *Package) CreateDatabaseCommand(db *Database) string {
	return `psql -U postgres -c "` + db.CreateCommand() + `"`
}

func (pkg *Package) CreateUserCommand(user *User) string {
	return `psql -U postgres -c "` + user.CreateCommand() + `"`
}

func (pkg *Package) UpstartExecCommand() cmd.Command {
	return cmd.WriteFile("/etc/init/postgres.conf", utils.MustRenderTemplate(upstart, pkg), "root", 0644)
}

const upstart = `
start on runlevel [2345]
stop on runlevel [!2345]
setuid {{ .User }}
exec {{ .InstallDir }}/bin/postgres -D {{ .DataDir }}
`

func (pkg *Package) url() string {
	return "http://ftp.postgresql.org/pub/source/v{{ .Version }}/postgresql-{{ .Version }}.tar.gz"
}

func (pkg *Package) installDir() string {
	return utils.MustRenderTemplate("/opt/postgresql-{{ .Version }}", pkg)
}
