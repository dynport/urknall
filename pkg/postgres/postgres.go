package postgres

import (
	"fmt"
	"strings"

	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/utils"
)

func New(version string) *Package {
	return &Package{Version: version}
}

type Package struct {
	Version string `urknall:"default=9.3.2"`
	DataDir string `urknall:"default=/data/postgres"`
}

func (p *Package) User() string {
	return "postgres"
}

func (pkg *Package) Package(r *urknall.Runlist) {
	r.Add(
		cmd.InstallPackages("build-essential", "openssl", "libssl-dev", "flex", "zlib1g-dev", "libxslt1-dev", "libxml2-dev", "python-dev", "libreadline-dev", "bison"),
		cmd.Mkdir("/opt/src/", "root", 0755),
		cmd.DownloadAndExtract(pkg.url(), "/opt/src/"),
		cmd.And(
			"cd /opt/src/postgresql-{{ .Version }}",
			"./configure --prefix="+pkg.InstallDir(),
			"make",
			"make install",
		),
		cmd.AddUser(pkg.User(), true),
	)
}

func (p *Package) InstallDir() string {
	return "/opt/postgresql-" + p.Version
}

func (pkg *Package) InitDbCommand() cmd.Command {
	return cmd.And(
		cmd.Mkdir(pkg.DataDir, pkg.User(), 0755),
		"su -l "+pkg.User()+" -c '"+pkg.InstallDir()+"/bin/initdb -D "+pkg.DataDir+" -E utf8 --auth-local=trust'",
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
	return pkg.InstallDir() + "/bin/" + `psql -U postgres -c "` + db.CreateCommand() + `"`
}

func (pkg *Package) CreateUserCommand(user *User) string {
	return pkg.InstallDir() + "/bin/" + `psql -U postgres -c "` + user.CreateCommand() + `"`
}

func (pkg *Package) InstallContribModule(module string) string {
	cmds := []string{
		"cd /opt/src/postgresql-" + pkg.Version + "/contrib/" + module,
		"make install",
		fmt.Sprintf(`%s/bin/psql -U postgres -c "CREATE EXTENSION %s"`, pkg.InstallDir(), module),
	}

	return strings.Join(cmds, " && ")
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
