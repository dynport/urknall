.PHONY: build check clean coverage default deps help test

ASSETS      := $(shell find assets -type f | grep -v ".go$$")
EXTRA_DEPS  := github.com/dynport/dgtk/goassets github.com/smartystreets/goconvey
DEPS        := $(shell go list ./... | xargs go list -f '{{join .Deps "\n"}}' | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' 2>/dev/null | sort | uniq | grep -v "github.com/dynport/zwo")
IGN_PKGS    := github.com/dynport/zwo/assets github.com/dynport/zwo/pkg% github.com/dynport/zwo/example
ALL_PKGS    := $(shell go list ./...)
PACKAGES    := $(filter-out $(IGN_PKGS),$(ALL_PKGS))

default: build

build: assets/assets.go
	@go install $(PACKAGES)

check:
	@which go > /dev/null || echo "go not installed"
	@which goassets > /dev/null || echo "go assets missing, call 'go get github.com/dynport/dgtk/goassets'"

clean:
	@rm -f assets/assets.go example/main

coverage: 
	$(eval $@_TMP := $(shell mktemp -t coverage))
	$(eval $@_COLLECT_TMP := $(shell mktemp -t coverage))
	@echo "mode: set" > $($@_COLLECT_TMP)
	@for package in $(PACKAGES); do \
		go test -coverprofile=$($@_TMP) $$package && awk 'NR > 1' $($@_TMP) >> $($@_COLLECT_TMP); \
		echo "\c" > $($@_TMP); \
	done
	@[ -s $($@_COLLECT_TMP) ] && go tool cover -func=$($@_COLLECT_TMP)
	@rm -f $($@_TMP) $($@_COLLECT_TMP)

deps:
	@for package in $(EXTRA_DEPS) $(DEPS); do \
		go get -t -u $$package; \
	done

help:
	@echo "make [target] ..."
	@echo "Targets:"
	@echo "  build:    Build the library."
	@echo "  clean:    Clean up all artifacts."
	@echo ""
	@echo "  check:    Verify all required tools are available."
	@echo "  deps:     Get, install or update all dependencies."
	@echo ""
	@echo "  test:     Run the tests."
	@echo "  coverage: Run the tests and print information on test coverage."
	@echo ""
	@echo "  help:     Print this message."

test: build
	@go vet $(PACKAGES)
	@go test -v $(PACKAGES)

assets/assets.go: $(ASSETS)
	@goassets ./assets > /dev/null 2>&1


