# portcheck Makefile
# Cross-platform development tasks

.PHONY: build build-all test lint clean install run help

# Version info (injected at build time)
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X github.com/uncaughtx/portcheck/internal/cli.Version=$(VERSION) -X github.com/uncaughtx/portcheck/internal/cli.Commit=$(COMMIT) -X github.com/uncaughtx/portcheck/internal/cli.BuildDate=$(DATE)"

# Binary names
BINARY_NAME := portcheck
BINARY_WINDOWS := $(BINARY_NAME).exe
BINARY_LINUX := $(BINARY_NAME)-linux
BINARY_DARWIN := $(BINARY_NAME)-darwin

# Default target
.DEFAULT_GOAL := help

## build: Build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/portcheck

## build-windows: Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_WINDOWS) ./cmd/portcheck

## build-linux: Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_LINUX) ./cmd/portcheck

## build-darwin: Build for macOS (Intel)
build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_DARWIN) ./cmd/portcheck

## build-darwin-arm: Build for macOS (Apple Silicon)
build-darwin-arm:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_DARWIN)-arm64 ./cmd/portcheck

## build-all: Build for all platforms
build-all: build-windows build-linux build-darwin build-darwin-arm
	@echo "Built binaries for all platforms"
	@ls -la portcheck*

## test: Run tests
test:
	go test -race -v ./...

## test-cover: Run tests with coverage
test-cover:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	golangci-lint run --timeout=5m

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## vet: Run go vet
vet:
	go vet ./...

## tidy: Tidy dependencies
tidy:
	go mod tidy

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_WINDOWS) $(BINARY_LINUX) $(BINARY_DARWIN) $(BINARY_DARWIN)-arm64
	rm -f coverage.out coverage.html

## install: Install locally
install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/

## run: Run the application
run: build
	./$(BINARY_NAME)

## deps: Download dependencies
deps:
	go mod download

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"

## version: Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

## help: Show this help
help:
	@echo "portcheck - Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
