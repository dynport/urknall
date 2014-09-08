// The urknall binary: see http://urknall.dynport.de/docs/binary/ for further information.
package main

import (
	"log"
	"os"

	"github.com/dynport/dgtk/cli"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

func main() {
	e := router().RunWithArgs()
	switch e {
	case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
	// ignore
	default:
		logger.Fatal(e)
	}
}
