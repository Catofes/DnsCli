all:
	mkdir -p build
	env CGO_ENABLED=0 go build -o build/dns -ldflags "-X 'main._version_=$(shell git describe --tags)'"\
                          	 github.com/Catofes/DnsCli/cmd/dns
test:
	go test github.com/Catofes/DnsCli/cmd/dns
