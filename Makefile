.PHONY: default build test vet

default: build test vet

build:
	@go get ./...

test:
	@go test $(PACKAGES)

vet:
	@go vet $(PACKAGES)

