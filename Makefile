.PHONY: build check clean default deps example test vet

ASSETS      := $(shell find assets -type f | grep -v ".go$$")

EXTRA_DEPS  := github.com/dynport/dgtk/goassets github.com/smartystreets/goconvey code.google.com/p/go.tools/cmd/vet github.com/dynport/gocli
DEPS        := $(shell go list ./... | xargs go list -f '{{join .Deps "\n"}}' | grep -e "$github.com\|code.google.com\|launchpad.net" | sort | uniq | grep -v "github.com/dynport/urknall")
IGN_PKGS    := github.com/dynport/urknall/assets
ALL_PKGS    := $(shell go list ./...)
PACKAGES    := $(filter-out $(IGN_PKGS),$(ALL_PKGS))
IGN_TEST_PKGS := github.com/dynport/urknall/pkg/%
TEST_PKGS   := $(filter-out $(IGN_TEST_PKGS),$(PACKAGES))

default: build

build: assets.go
	@go install $(PACKAGES)

check:
	@which go > /dev/null || echo "go not installed"
	@which goassets > /dev/null || echo "go assets missing, call 'go get github.com/dynport/dgtk/goassets'"

clean:
	@rm -f assets.go example/main

deps:
	@for package in $(EXTRA_DEPS) $(DEPS); do \
		echo "Installing $$package"; \
		go get -u $$package; \
	done

example: build
	@go run example/main.go

test: build
	@go test $(TEST_PKGS)

vet: build
	@go vet $(PACKAGES)

assets.go: $(ASSETS)
	@rm -f $@
	@goassets assets > /dev/null 2>&1

