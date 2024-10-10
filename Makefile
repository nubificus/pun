
default: pun

prepare:
	go mod tidy
	go mod vendor

pun: prepare
	CGO_ENABLED=0 go build -o $@ --ldflags "-s -w"

clean:
	rm -rf pun
