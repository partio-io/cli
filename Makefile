BINARY=partio
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: build run install test lint clean demo demo-raw

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

demo:
	vhs demo.tape
	gifsicle assets/demo.gif "#0-539" -d1 "#540-" -o assets/demo.gif
	gifsicle --unoptimize assets/demo.gif --delete "#420-469" "#470-519" "#520-559" -O2 -o assets/demo.gif

demo-raw:
	vhs demo.tape

.DEFAULT_GOAL := build
