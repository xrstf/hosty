default: build
run: build fire

build:
	go build -v -tags netgo -ldflags '-s -w' .

fire:
	./hosty serve config.yaml

package: build
	rm -f package.tar.gz
	tar czf package.tar.gz hosty resources www LICENSE.md README.md config.yaml.dist
