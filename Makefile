.PHONY: build test lint clean

BINARY := microwave
VERSION ?= dev

build:
	go build -ldflags "-s -w -X github.com/microwave-sh/microwave-cli/internal/version.Version=$(VERSION)" -o $(BINARY) .

test:
	go test ./... -v

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)
	rm -rf dist/
