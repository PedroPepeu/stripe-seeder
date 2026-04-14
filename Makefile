BINARY  = stripe-seeder
MODULE  = github.com/PedroPepeu/stripe-seeder
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION)"
DIST    = dist

.PHONY: build install clean release test

## build: compile binary for the current platform
build:
	go build $(LDFLAGS) -o $(BINARY) .

## install: install binary to $GOPATH/bin (enables: go install)
install:
	go install $(LDFLAGS) .

## test: run all tests
test:
	go test ./...

## release: cross-compile binaries for all platforms into dist/
release: clean test
	mkdir -p $(DIST)
	GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-amd64     .
	GOOS=linux   GOARCH=arm64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-arm64     .
	GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-amd64    .
	GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-arm64    .
	GOOS=windows GOARCH=amd64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-windows-amd64.exe .
	@echo ""
	@echo "Binaries in $(DIST)/:"
	@ls -lh $(DIST)/

## clean: remove compiled binary and dist/
clean:
	rm -rf $(BINARY) $(DIST)

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
