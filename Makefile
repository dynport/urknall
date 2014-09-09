.PHONY: build deps test vet

EXTRA_DEPS  := github.com/smartystreets/goconvey github.com/jacobsa/oglematchers
DEPS        := $(shell go list ./... | xargs go list -f '{{join .Deps "\n"}}' | grep -e "$github.com\|code.google.com\|launchpad.net" | sort | uniq | grep -v "github.com/dynport/urknall")
IGN_PKGS    := github.com/dynport/urknall/examples
ALL_PKGS    := $(shell go list ./...)
PACKAGES    := $(filter-out $(IGN_PKGS),$(ALL_PKGS))

build:
	go get github.com/dynport/urknall/urknall

deps:
	@for package in $(EXTRA_DEPS) $(DEPS); do \
		echo "Installing $$package"; \
		go get -u $$package; \
	done

test:
	@go test $(PACKAGES)

vet:
	@go vet $(PACKAGES)

