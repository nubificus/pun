COMMIT := $(shell git describe --dirty --long --always)
VERSION := $(shell cat ./VERSION)
VERSION := $(VERSION)-$(COMMIT)
ARCH := $(shell dpkg --print-architecture)

default: build ;

prepare:
	@go mod tidy
	@mkdir -p dist

build: prepare
	@GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w" -ldflags "-w" -ldflags "-linkmode 'external' -extldflags '-static'" \
          -ldflags "-X main.version=${VERSION}" -o ./dist/pun_${ARCH} ./

install:
	@mv ./dist/pun_${ARCH} /usr/local/bin/pun

uninstall:
	@rm -f /usr/local/bin/pun

clean:
	@rm -fr ./dist/
	@rm -f ./Tempfile

build_aarch64: prepare
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -ldflags "-w" -ldflags "-linkmode 'external' -extldflags '-static'" \
          -ldflags "-X main.version=${VERSION}" -o ./dist/pun_aarch64 ./

build_amd64: prepare
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -ldflags "-w" -ldflags "-linkmode 'external' -extldflags '-static'" \
          -ldflags "-X main.version=${VERSION}" -o ./dist/pun_amd64 ./

all: build_aarch64 build_amd64
