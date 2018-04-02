default: build
run: build fire

build: fix
	go build -v -tags netgo -ldflags '-s -w' .

fix: *.go
	goimports -l -w .
	gofmt -l -w -s .

deps:
	go get github.com/gin-gonic/gin
	go get github.com/jmoiron/sqlx
	go get github.com/kardianos/osext
	go get github.com/mattn/go-sqlite3
	go get github.com/rainycape/unidecode
	go get golang.org/x/crypto/bcrypt
	go get golang.org/x/tools/cmd/goimports

fire:
	./hosty serve config.yaml

package: build
	rm -f package.tar.gz
	tar czf package.tar.gz hosty resources www LICENSE.md README.md config.yaml.dist
