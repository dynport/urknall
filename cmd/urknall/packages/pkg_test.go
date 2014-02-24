package main

import (
	"testing"

	"github.com/dynport/urknall"
)

func TestPackages(t *testing.T) {
	packages := []urknall.Package{
		&Nginx{},
		&Redis{},
		&Postgres{},
		&Ruby{},
		&SyslogNg{},
		&ElasticSearch{},
		&OpenVPN{},
		&HAProxy{},
		&Golang{},
	}
	if len(packages) < 1 {
		t.Errorf("expected > 0 packages, got %d", len(packages))
	}
}
