.PHONY: default build test clean

default: build

ASSETS := $(shell find assets -type f | grep -v ".go$$")
assets/assets.go: $(ASSETS)
	@goassets ./assets

build: assets/assets.go
	@go build ./...

PACKAGES := $(shell go list ./...)
test: build
	@go vet ./...
	@for package in $(PACKAGES); do \
		export TMP=$$(mktemp -t "zwo_tc"); \
		go test -coverprofile=$$TMP $$package; \
		[[ -s $$TMP ]] && go tool cover -func=$$TMP; \
		unlink $$TMP; \
		unset TMP; \
	done

clean:
	@rm -f assets/assets.go

