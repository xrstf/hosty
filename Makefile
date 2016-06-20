default: build
run: build fire

build: fix
	go build -v .

fix: *.go
	goimports -l -w .
	gofmt -l -w -s .

fire:
	./hosty serve config.yaml

package: build
	rm -f package.tar.gz
	tar czf package.tar.gz hosty resources www LICENSE.md README.md config.yaml.dist
