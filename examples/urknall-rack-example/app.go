package main

import (
	"github.com/dynport/urknall"
)

// App is a urknall.Template which will be used to be executed on a target
//
// this template can hold Variables (e.g. RubyVersion)
type App struct {
	RubyVersion string `urknall:"required=true"`
	User        string `urknall:"required=true"`
}

func (tpl *App) Render(p urknall.Package) {
	// update all system packages (once)
	p.AddCommands("update", UpdatePackages())

	// install some often known system packages
	p.AddCommands("packages",
		InstallPackages("vim-nox", "tmux", "ntp", "htop"),
	)

	// install nginx from ubuntu package
	p.AddCommands("nginx", InstallPackages("nginx"))

	// install ruby from source
	p.AddTemplate("ruby", &Ruby{Version: tpl.RubyVersion})

	// false: no system user
	p.AddCommands("app.user", AddUser(tpl.User, false))

	// write user profile from appProfile constant
	p.AddCommands("app.profile",
		WriteFile("/home/app/.profile", appProfile, tpl.User, 0644),
	)

	// executes the command (gem install puma...) as user app
	p.AddCommands("app.gems",
		AsUser(tpl.User, "gem install puma --no-ri --no-rdoc"),
	)
	p.AddCommands("app.code",
		WriteFile("/home/app/config.ru", configRu, tpl.User, 0644),
	)
	p.AddCommands("app.upstart",
		WriteFile("/etc/init/app.conf", appUpstart, "root", 0644),
	)
	p.AddCommands("app.start",
		Shell("start app"),
	)

	// all statements inside e.g. an AddCommands call are cached by statements
	// every time a statement changes all statements starting from that statement (including that statement) are also executed
	// Example: if the content of appNginx changes, nginx -t and either reload or start with be executed again
	p.AddCommands("app.nginx",
		WriteFile("/etc/nginx/sites-available/default", appNginx, "root", 0644),
		Shell("/usr/sbin/nginx -t"),
		Shell("if /etc/init.d/nginx status; then /etc/init.d/nginx reload; else /etc/init.d/nginx start; fi"),
	)
}

// TEMPLATES
//
// the easiest way to work with templates is to have them inline (in e.g. constants).
// Other options are tools like github.com/dynport/dgtk/goassets, go-bindata, etc
const appNginx = `upstream puma {
  server unix:/home/{{ .User }}/puma.socket fail_timeout=0;
}

server {
    listen 80;
    server_name _;

	root /home/{{ .User }}/public;

    location @puma {
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header Host $http_host;
      proxy_redirect off;
      proxy_pass http://puma;
  	}

	try_files $uri @puma;
}
`

// in file templates the variables and methods of the urknall.Template
// can also be used for "rendering"
const appUpstart = `
script
su {{ .User }} <<"EOF"
cd /home/{{ .User }}
source .profile
puma -b unix:///home/app/puma.socket -S /home/app/puma.state
EOF
end script
`

const appProfile = `
export GEM_HOME=$HOME/gems
export GEM_ROOT=$HOME/gems
export PATH=$GEM_HOME/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games
`

const configRu = `#!/bin/env ruby
app = lambda do |env|
  [200, {"Content-Type" => "text/plain"}, ["Hello World!"]]
end

run app
`
