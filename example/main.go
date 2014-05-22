package main

import (
	"log"
	"os"

	"github.com/dynport/urknall"
)

// This small example will build a target with redis and nginx. A pretty
// pointless setup, but good enough to demonstrate features.
//
// In urknall we will build a target with a package. The machine
// that is to be set up automatically is called target. The package is a specification of what
// should be set up and how. This is a three layer architecture:
// - commands are the atoms, the stuff actually executed on the target (think
//   of it as shell commands).
// - a task is an ordered list of commands. It has a caching mechanism
//   built in that determines for any command whether it has been
//   executed previously (in that case it won't be executed again). This
//   prevents the system from doing stuff over and over again. If a
//   command changes, all following commands of the ordered list of commands will be executed
//   again.
// - a package is a container of tasks and other packages. It is used to
//   structure projects.
// Neither tasks nor packages are generate by the user. Instead users are encouraged to
// create PackageBuilders and TaskBuilders that add things to given tasks or packages (or targets ?).

func main() {
	// First set up some logging. Stdout is fine in most cases, but it also could be
	// sent to a file or somewhere else.
	defer urknall.OpenLogger(os.Stdout).Close()

	// Next define the target.
	target, e := urknall.NewSshTargetWithPassword("ubuntu@192.168.56.10", "ubuntu")
	if e != nil {
		panic(e.Error())
	}

	// Create an object of a type that implements the
	// PackageBuilder interface.
	pkgBuilder := &Example{Hostname: "example"}

	// Run the PackageBuilder on the target to build the package's content
	e = urknall.Run(target, pkgBuilder)
	if e != nil {
		log.Fatal(e)
	}
}

// The example package builder. It has no configuration, but implements
// the PackageBuilder interface.
type Example struct {
	Hostname string `urknall:"required=true"`
}

// The BuildPackage method of the PackageBuilder interface is given a Package
// (from urknall) that can be populated with Tasks and other Packages. Tasks or Packages
// are added with an identifying name. The names are concatenated in
// the resulting hierarchy (note the first column of the logged output).
// Tasks can either be created without configuration using the
// NewTask function or with configuration as in System-Task example.
func (ex *Example) Render(pkg urknall.Package) {
	pkg.AddCommands("update", UpdatePackages())
	pkg.AddTemplate("hostname", &System{Hostname: ex.Hostname})
	pkg.AddTemplate("srv", &Services{
		NginxVersion: "1.4.4",
		RedisVersion: "2.8.9",
	})
}

// A simple task builder creating a task that will set system properties (the
// hostname in this example).
type System struct {
	Hostname string `urknall:"required=true"`
}

// Task builders must implement the TaskBuilder interface and thereby implement the
// BuildTask method. It is called with a Task to which commands can be added similarly
// as in the case of the Package above.
func (sys *System) Render(task urknall.Package) {
	task.AddCommands("base",
		Shell("hostname localhost"), // Set hostname to make sudo happy.
		&FileCommand{Path: "/etc/hostname", Content: sys.Hostname},
		&FileCommand{Path: "/etc/hosts", Content: "127.0.0.1 {{ .Hostname }} localhost"},
		Shell("hostname -F /etc/hostname"),
	)
}

// Another PackageBuilder that now uses Task- and PackageBuilders predefined in
// urknall packages (see urknall binary for more information).
type Services struct {
	NginxVersion string `urknall:"default='1.4.4'"`
	RedisVersion string `urknall:"default='2.8.8'"`
}

func (srv *Services) Render(pkg urknall.Package) {
	pkg.AddTemplate("nginx", &Nginx{Version: srv.NginxVersion})
	pkg.AddTemplate("redis", &Redis{Version: srv.RedisVersion})
}
