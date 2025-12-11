# Gmail CLI Makefile

BINARY_NAME=gmail-cli
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/bentsolheim/gmail-cli/pkg/version.Version=$(VERSION) -X github.com/bentsolheim/gmail-cli/pkg/version.Commit=$(COMMIT) -X github.com/bentsolheim/gmail-cli/pkg/version.BuildDate=$(BUILD_DATE)"

.PHONY: build install clean test lint

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/gmail-cli/

install:
	go install $(LDFLAGS) ./cmd/gmail-cli/

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

test:
	go test -v ./...

lint:
	golangci-lint run

# Cross-compilation targets
.PHONY: build-all
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/gmail-cli/
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/gmail-cli/

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/gmail-cli/
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/gmail-cli/

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/gmail-cli/
