package main

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/logging"
	"log"
	"os"
	"strconv"
	"time"
)

func init() {
	log.SetFlags(0)
}

func initHost() *urknall.Host {
	ip := os.Getenv("TEST_IP")
	if ip == "" {
		log.Fatal("TEST_IP must be set")
	}
	host := &urknall.Host{IP: ip, User: os.Getenv("TEST_USER")}
	host.AddPackage("hello_world", urknall.NewPackage("echo hello world", "echo "+strconv.FormatInt(time.Now().Unix(), 10)))
	return host
}

func main() {
	host := initHost()
	logger := &logging.StdoutLogger{}
	defer logger.Close()
	logger.Formatter = logger.DefaultFormatter
	e := logger.Start()
	if e != nil {
		log.Fatal(e.Error())
	}
	e = host.Provision(nil)
	if e != nil {
		log.Fatal(e.Error())
	}
}
