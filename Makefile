BINARY=partio
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: build run install test lint clean

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/partio

run: build
	./$(BINARY) $(ARGS)

install:
	go install $(LDFLAGS) ./cmd/partio

test:
	go test ./... -v

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)

.DEFAULT_GOAL := build
