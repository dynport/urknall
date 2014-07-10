package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dynport/urknall"
)

var logger = log.New(os.Stderr, "", 0)

// execute with	urknall-rack-example -l <login> -H <host> -p <password>
func main() {
	if e := run(); e != nil {
		logger.Printf("ERROR: " + e.Error())
		flag.Usage()
	}
}

func run() error {
	// Setup logging to stdout
	// all logs are published to a pubsub system and urknall.OpenLogger adds a custom consumer
	// which writes to the provided io.Writer (in that case os.Stdout)
	// the logger needs to be closed to flush pending logs
	defer urknall.OpenLogger(os.Stdout).Close()

	// get urknall.Target from flags
	target, e := targetFromFlags()
	if e != nil {
		return e
	}
	// Execute a urknall.Template (App) on the provided target
	return urknall.Run(target, &App{RubyVersion: "2.1.2", User: "app"})
}

var (
	host     = flag.String("H", "", "SSH host (required)")
	login    = flag.String("l", "", "SSH login")
	password = flag.String("p", "", "SSH password")
)

func targetFromFlags() (urknall.Target, error) {
	flag.Parse()

	if *host == "" {
		return nil, fmt.Errorf("no host provided")
	}
	creds := *host

	if *login != "" {
		creds = *login + "@" + creds
	}
	logger.Printf("using ssh %q", creds)

	if *password != "" {
		// urknall.NewSshTargetWithPassword creates a target with password authentication
		// use <login>@<host> to specify ssh login and host.
		// only use "<host>" to use the defaults (either your local user or from $HOME/.ssh/config
		return urknall.NewSshTargetWithPassword(creds, *password)
	}

	// use urknall.NewSshTarget to use public key (using local ssh-agent) authentication
	return urknall.NewSshTarget(creds)
}
