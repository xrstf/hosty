default: build
run: build fire

build: fix
	go build -v .

fix: *.go
	goimports -l -w .
	gofmt -l -w .

fire:
	./raziel.exe --config config.json
