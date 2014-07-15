package main

import "github.com/dynport/dgtk/cli"

func router() *cli.Router {
	router := cli.NewRouter()
	router.Register("init", &Init{}, "Initialize")
	router.Register("packages/add", &AddPackage{}, "Add Package")
	router.Register("packages/list", &PackagesList{}, "List Packages")
	return router
}
