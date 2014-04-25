// The urknall binary used to bootstrap, create, and maintain urknall projects.
//
// While the urknall library has the basic functionality to manage the provisioning this binary helps to prevent some
// major problems of the library. Initially the library contained the commands and packages that were imported directly.
// This approach had some drawbacks. First a change in the core library would trigger reprovisioning for all users,
// which could have desastrous effects (think of non-idempotent commands like creating a database). Second this feels
// like giving advice on how to do stuff. This is not it. This is only our way of doing things. We'd like to provide
// those as starting point, but users should be able to adopt that to their likings.
//
// This binary will create a basic project on `init` and allows to add required packages for many services. Updating the
// templates is also possible (but care should be taken that a version control system is used, so that nothing can be
// overwritten).
package main

import (
	"log"
	"os"

	"github.com/dynport/dgtk/cli"
)

var (
	router = cli.NewRouter()
	logger = log.New(os.Stderr, "", 0)
)

func main() {
	e := router.RunWithArgs()
	switch e {
	case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
	//
	default:
		logger.Fatal(e)
	}
}

func init() {
	wd, e := os.Getwd()
	if e != nil {
		logger.Fatal(e)
	}
	base := base{BaseDir: wd}
	router.Register("init", &Init{base: base}, "Initialize")
	router.Register("packages/add", &AddPackage{base: base}, "Add Package")
	router.Register("packages/list", &PackagesList{}, "List Packages")
}
