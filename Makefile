default: build
run: build fire

build: fix
	echo package main > version.go
	git log -n 1 --format="const version = \"%h\"" >> version.go
	go build -v .

fix: *.go
	goimports -l -w .
	gofmt -l -w .

fire:
	./hosty serve config.yaml

package: build
	rm -f package.tar.gz
	tar czf package.tar.gz hosty resources www LICENSE.md README.md config.yaml.dist
