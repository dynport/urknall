package main

import "github.com/dynport/urknall"

type Postgres struct {
	Version string `urknall:"required=true"` // 9.3.4
	DataDir string `urknall:"default=/data/postgres"`
	User    string `urknall:"default=postgres"`

	InitDb bool
}

func (pgres *Postgres) Render(pkg urknall.Package) {
	pkg.AddCommands("packages",
		InstallPackages("build-essential", "openssl", "libssl-dev", "flex", "zlib1g-dev", "libxslt1-dev", "libxml2-dev", "python-dev", "libreadline-dev", "bison"),
	)
	pkg.AddCommands("download", DownloadAndExtract("{{ .Url }}", "/opt/src/"))
	pkg.AddCommands("user", AddUser("{{ .User }}", true))
	pkg.AddCommands("build",
		And(
			"cd /opt/src/postgresql-{{ .Version }}",
			"./configure --prefix={{ .InstallDir }}",
			"make",
			"make install",
		),
	)
	pkg.AddCommands("upstart",
		WriteFile("/etc/init/postgres.conf", postgresUpstart, "root", 0644),
	)
	if pgres.InitDb {
		user := pgres.User
		if user == "" {
			user = "postgres"
		}
		pkg.AddCommands("init_db",
			Mkdir(pgres.DataDir, "{{ .User }}", 0755),
			AsUser(user, "{{ .InstallDir }}/bin/initdb -D {{ .DataDir }} -E utf8 --auth-local=trust"),
		)
	}
}

func (pgres *Postgres) InstallDir() string {
	if pgres.Version == "" {
		panic("Version must be set")
	}
	return "/opt/postgresql-" + pgres.Version
}

// some helpers for e.g. database creation
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

func (pgres *Postgres) CreateDatabaseCommand(db *PostgresDatabase) string {
	return pgres.InstallDir() + "/bin/" + `psql -U postgres -c "` + db.CreateCommand() + `"`
}

func (pgres *Postgres) CreateUserCommand(user *PostgresUser) string {
	return pgres.InstallDir() + "/bin/" + `psql -U postgres -c "` + user.CreateCommand() + `"`
}

const postgresUpstart = `
start on runlevel [2345]
stop on runlevel [!2345]
setuid {{ .User }}
exec {{ .InstallDir }}/bin/postgres -D {{ .DataDir }}
`

func (pgres *Postgres) Url() string {
	return "http://ftp.postgresql.org/pub/source/v{{ .Version }}/postgresql-{{ .Version }}.tar.gz"
}
