package main

import "github.com/dynport/dgtk/cli"

func router() *cli.Router {
	router := cli.NewRouter()
	router.Register("init", &initProject{}, "Initialize a basic urknall project.")
	router.Register("templates/add", &templatesAdd{}, "Add templates to project.")
	router.Register("templates/list", &templatesList{}, "List all available templates.")
	return router
}
