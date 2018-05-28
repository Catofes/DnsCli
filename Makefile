all:
	mkdir -p build
	go build -o build/dns -ldflags "-X 'main._version_=$(shell git log --pretty=format:"%h" -1)'"\
                          	 github.com/Catofes/DnsCli
test:
	env CONFIG_PATH=../config/config.json go test github.com/Catofes/DnsCli
