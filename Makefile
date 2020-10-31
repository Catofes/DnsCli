all:
	mkdir -p build
	env CGO_ENABLE=0 go build -o build/dns -ldflags "-X 'main._version_=$(shell git log --pretty=format:"%h" -1)'"\
                          	 github.com/Catofes/DnsCli/cmd/dns
test:
	go test github.com/Catofes/DnsCli/cmd/dns
