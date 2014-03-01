.PHONY: deps test vet

ASSETS      := $(shell find urknall/packages -type f | grep -v ".go$$")

EXTRA_DEPS  := github.com/dynport/dgtk/goassets github.com/smartystreets/goconvey github.com/jacobsa/oglematchers
DEPS        := $(shell go list ./... | xargs go list -f '{{join .Deps "\n"}}' | grep -e "$github.com\|code.google.com\|launchpad.net" | sort | uniq | grep -v "github.com/dynport/urknall")
IGN_PKGS    := github.com/dynport/urknall/urknall/packages
ALL_PKGS    := $(shell go list ./...)
PACKAGES    := $(filter-out $(IGN_PKGS),$(ALL_PKGS))

${GOPATH}/bin/urknall: urknall/assets.go urknall/*.go
	go install github.com/dynport/urknall/urknall

deps:
	@for package in $(EXTRA_DEPS) $(DEPS); do \
		echo "Installing $$package"; \
		go get -u $$package; \
	done

test:
	@go test $(PACKAGES)

vet:
	@go vet $(PACKAGES)

urknall/assets.go: $(ASSETS)
	rm -f $@
	cd urknall && goassets assets > /dev/null 2>&1

