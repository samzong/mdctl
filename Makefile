BINARY=mdctl
VERSION=$(shell git describe --tags || echo "unknown version")
BUILDTIME=$(shell date -u)
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags '-X "github.com/samzong/mdctl/cmd.Version=$(VERSION)" -X "github.com/samzong/mdctl/cmd.BuildTime=$(BUILDTIME)"'

# Homebrew related variables
CLEAN_VERSION=$(shell echo $(VERSION) | sed 's/^v//')
DOWNLOAD_URL=https://github.com/samzong/mdctl/releases/download/$(VERSION)/mdctl-$(CLEAN_VERSION)-darwin-amd64.tar.gz
HOMEBREW_TAP_REPO=homebrew-tap
FORMULA_FILE=Formula/mdctl.rb
BRANCH_NAME=update-mdctl-$(CLEAN_VERSION)

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

.PHONY: update-homebrew
update-homebrew:
	@if [ -z "$(GH_PAT)" ]; then \
		echo "Error: GH_PAT environment variable is not set"; \
		exit 1; \
	fi
	@echo "Updating Homebrew formula for version $(CLEAN_VERSION)"
	@rm -rf tmp && mkdir -p tmp
	@cd tmp && \
		curl -L -o mdctl.tar.gz $(DOWNLOAD_URL) && \
		SHA256=$$(shasum -a 256 mdctl.tar.gz | cut -d ' ' -f 1) && \
		git clone https://$(GH_PAT)@github.com/samzong/$(HOMEBREW_TAP_REPO).git && \
		cd $(HOMEBREW_TAP_REPO) && \
		git checkout -b $(BRANCH_NAME) && \
		sed -i '' \
			-e 's/version "[^"]*"/version "$(CLEAN_VERSION)"/' \
			-e 's/sha256 "[^"]*"/sha256 "'$$SHA256'"/' \
			$(FORMULA_FILE) && \
		git config user.name "GitHub Actions" && \
		git config user.email "github-actions@github.com" && \
		git add $(FORMULA_FILE) && \
		git commit -m "chore: update mdctl to $(VERSION)" && \
		git push -u origin $(BRANCH_NAME) && \
		curl -X POST \
			-H "Accept: application/vnd.github.v3+json" \
			-H "Authorization: token $(GH_PAT)" \
			https://api.github.com/repos/samzong/$(HOMEBREW_TAP_REPO)/pulls \
			-d '{"title":"chore: update mdctl to $(VERSION)","body":"Update mdctl to $(VERSION)\n\n- Version: $(CLEAN_VERSION)\n- SHA256: '$$SHA256'","head":"$(BRANCH_NAME)","base":"main"}'
	@rm -rf tmp

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
	@echo "  update-homebrew: Update Homebrew formula (requires GH_PAT)"

.DEFAULT_GOAL := help
