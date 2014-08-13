package main

import "github.com/dynport/dgtk/cli"

func router() *cli.Router {
	router := cli.NewRouter()
	router.Register("init", &Init{}, "Initialize")
	router.Register("templates/add", &templatesAdd{}, "Add templates")
	router.Register("templates/list", &templatesList{}, "List templates")
	return router
}
