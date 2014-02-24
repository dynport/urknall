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
