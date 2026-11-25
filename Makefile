BINARY=mdctl
VERSION=$(shell git describe --tags || echo "unknown version")
BUILDTIME=$(shell date -u)
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags '-X "github.com/samzong/mdctl/cmd.Version=$(VERSION)" -X "github.com/samzong/mdctl/cmd.BuildTime=$(BUILDTIME)"'

.PHONY: deps
deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod verify

.PHONY: build
build: deps
	$(GOBUILD) -o bin/$(BINARY)

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -rf bin/
	go clean -i ./...

.PHONY: fmt
fmt:
	go fmt ./...
	go mod tidy

.PHONY: all
all: clean fmt build test

.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  deps: Install Go dependencies"
	@echo "  build: Build the binary"
	@echo "  test: Run tests"
	@echo "  clean: Clean up build artifacts"
	@echo "  fmt: Format the code"
	@echo "  all: Clean, format, build, and test"

.DEFAULT_GOAL := help
