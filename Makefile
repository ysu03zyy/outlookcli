PREFIX ?= $(shell go env GOPATH)
BINDIR ?= $(PREFIX)/bin
VERSION ?= dev

.PHONY: install build clean release-all

install:
	go install -ldflags "-X main.version=$(VERSION)" ./cmd/outlookcli

build:
	mkdir -p bin
	go build -ldflags "-X main.version=$(VERSION)" -o bin/outlookcli ./cmd/outlookcli

clean:
	rm -rf bin/

# Cross-compile release binaries (run from macOS or Linux with Go installed)
release-all: clean
	mkdir -p bin
	GOOS=darwin  GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o bin/outlookcli-darwin-amd64 ./cmd/outlookcli
	GOOS=darwin  GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o bin/outlookcli-darwin-arm64 ./cmd/outlookcli
	GOOS=linux   GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o bin/outlookcli-linux-amd64 ./cmd/outlookcli
	GOOS=linux   GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o bin/outlookcli-linux-arm64 ./cmd/outlookcli
