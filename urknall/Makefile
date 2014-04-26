.PHONY: assets

build: assets
	go get ./...

assets:
	rm -f assets.go packages/*.test
	goassets packages
