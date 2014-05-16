package main

import (
	"log"
	"os"

	"github.com/dynport/urknall"
)

// This small example will build a target with redis and nginx. A pretty
// pointless setup, but good enough to demonstrate features.
//
// In urknall will build a target with a package. The target is the machine
// that should be set up automatically. The package is a specification of what
// should be set up and how. This is a three layer architecture:
// - commands are the atoms, the stuff actually executed on the target (think
//   of it as shell commands).
// - a task is a list of commands with order kept. It has a caching mechanism
//   built in, that will determine for each command whether it has been
//   executed previously (in that case it won't be executed again). This
//   prevents the system from repeatedly doing stuff over and over again. If a
//   command changed it and all following commands of the task will be executed
//   again.
// - a package is a container of tasks and other packages. It is used to
//   give projects structure.
// Neither tasks nor packages are generate by the user, Instead users will
// create Package- and TaskBuilders that are given an respective item for
// adding things.
func main() {
	// First set up some logigng. Stdout is fine in most cases, but could be
	// sent to a file or somewhere else.
	defer urknall.OpenLogger(os.Stdout).Close()

	// Next define what the target is.
	target, e := urknall.NewSshTargetWithPassword("ubuntu@192.168.56.10", "ubuntu")
	if e != nil {
		panic(e.Error())
	}

	// Create some package builder, aka something that implements the
	// PackageBuilder interface.
	pkgBuilder := &Example{Hostname: "example"}

	// Build the package and apply it to the target.
	e = urknall.Run(target, pkgBuilder)
	if e != nil {
		log.Fatal(e)
	}
}

// The example package builder. This one has no configuration, but implements
// the PackageBuilder interface.
type Example struct {
	Hostname string `urknall:"required=true"`
}

// The BuildPackage method of the PackageBuilder interface is given an package
// (from urknall) that can be populated with tasks and subpackages. The names
// are given as first parameter to the package's Add method are concatenated in
// the resulting hierarchy (note the first column of the logged output).
// Tasks can either be created ad hoc, i.e. without configuration using the
// NewTask function, or with configuration as in System task example.
func (ex *Example) BuildPackage(pkg urknall.Package) {
	pkg.Add("update", urknall.NewTask(UpdatePackages()))
	pkg.Add("hostname", &System{Hostname: ex.Hostname})
	pkg.Add("srv", &Services{
		NginxVersion: "1.4.4",
		RedisVersion: "2.8.9",
	})
}

// A simple task builder creating a task that will set system properties (the
// hostname in this example).
type System struct {
	Hostname string `urknall:"required=true"`
}

// Task builders must implement the TaskBuilder interface, i.e. have the
// BuildTask method. It is given a task that can be populated with commands.
func (sys *System) BuildTask(task urknall.Task) {
	task.Add(
		"hostname localhost", // Set hostname to make sudo happy.
		&FileCommand{Path: "/etc/hostname", Content: sys.Hostname},
		&FileCommand{Path: "/etc/hosts", Content: "127.0.0.1 {{ .Hostname }} localhost"},
		"hostname -F /etc/hostname",
	)
}

// Another PackageBuilder that now uses Task- and PackageBuilders predefined in
// urknall packages (see urknall binary for more information).
type Services struct {
	NginxVersion string `urknall:"default='1.4.4'"`
	RedisVersion string `urknall:"default='2.8.8'"`
}

func (srv *Services) BuildPackage(pkg urknall.Package) {
	pkg.Add("nginx", &Nginx{Version: srv.NginxVersion})
	pkg.Add("redis", &Redis{Version: srv.RedisVersion})
}
