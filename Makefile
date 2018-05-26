all:
	mkdir -p build
	go build -o build/web -ldflags "-X 'dnscli._version_=$(shell git log --pretty=format:"%h" -1)'"\
                          	 github.com/Catofes/DnsCli
test:
	env CONFIG_PATH=../config/config.json go test gitlab.threebody.org/qiaohao/Backend-Web/web
