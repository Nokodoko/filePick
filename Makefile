.PHONY: build test lint install clean

BINARY := fp
VERSION := $(shell grep 'const Version' main.go | cut -d'"' -f2)

build:
	go build -o $(BINARY) .

test:
	go test ./... -v

lint:
	golangci-lint run

install: build
	cp $(BINARY) $(GOPATH)/bin/

clean:
	rm -f $(BINARY)
