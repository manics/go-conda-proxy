# VERSION := $(shell git describe --tags --always HEAD)
# GOFLAGS = -ldflags "-X main.Version=$(VERSION)"

default: all

all: build test lint

lint:
	golangci-lint run

build:
	go build $(GOFLAGS) -o conda-parser ./cmd/conda-parser
	go build $(GOFLAGS) -o conda-proxy ./cmd/conda-proxy

test:
	go test -v ./...

clean:
	rm -f conda-parser conda-proxy

update-deps:
	go get -t -u ./...
	go mod tidy
