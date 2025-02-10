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
		git clone https://$(GH_PAT)@github.com/samzong/$(HOMEBREW_TAP_REPO).git && \
		cd $(HOMEBREW_TAP_REPO) && \
		git checkout -b $(BRANCH_NAME) && \
		cd .. && \
		for arch in "darwin-amd64" "darwin-arm64" "linux-amd64" "linux-arm64"; do \
			echo "Downloading $$arch package..." && \
			curl -L -o "mdctl-$$arch.tar.gz" "https://github.com/samzong/mdctl/releases/download/$(VERSION)/mdctl-$(CLEAN_VERSION)-$$arch.tar.gz" || exit 1; \
		done && \
		echo "Calculating SHA256 values..." && \
		SHA256_DARWIN_AMD64=$$(shasum -a 256 mdctl-darwin-amd64.tar.gz | cut -d ' ' -f 1) && \
		SHA256_DARWIN_ARM64=$$(shasum -a 256 mdctl-darwin-arm64.tar.gz | cut -d ' ' -f 1) && \
		SHA256_LINUX_AMD64=$$(shasum -a 256 mdctl-linux-amd64.tar.gz | cut -d ' ' -f 1) && \
		SHA256_LINUX_ARM64=$$(shasum -a 256 mdctl-linux-arm64.tar.gz | cut -d ' ' -f 1) && \
		echo "Verifying SHA256 values are different..." && \
		if [ "$$SHA256_DARWIN_AMD64" = "$$SHA256_DARWIN_ARM64" ] || \
		   [ "$$SHA256_DARWIN_AMD64" = "$$SHA256_LINUX_AMD64" ] || \
		   [ "$$SHA256_DARWIN_AMD64" = "$$SHA256_LINUX_ARM64" ] || \
		   [ "$$SHA256_DARWIN_ARM64" = "$$SHA256_LINUX_AMD64" ] || \
		   [ "$$SHA256_DARWIN_ARM64" = "$$SHA256_LINUX_ARM64" ] || \
		   [ "$$SHA256_LINUX_AMD64" = "$$SHA256_LINUX_ARM64" ]; then \
			echo "Error: Some SHA256 values are identical. This indicates a potential error in the build process." && \
			exit 1; \
		fi && \
		cd $(HOMEBREW_TAP_REPO) && \
		echo "Updating formula file..." && \
		sed -i '' \
			-e 's/version "[^"]*"/version "$(CLEAN_VERSION)"/' \
			-e 's/sha256 cellar: :any_skip_relocation, arm64_sonoma: "[^"]*"/sha256 cellar: :any_skip_relocation, arm64_sonoma: "'$$SHA256_DARWIN_ARM64'"/' \
			-e 's/sha256 cellar: :any_skip_relocation, arm64_ventura: "[^"]*"/sha256 cellar: :any_skip_relocation, arm64_ventura: "'$$SHA256_DARWIN_ARM64'"/' \
			-e 's/sha256 cellar: :any_skip_relocation, arm64_monterey: "[^"]*"/sha256 cellar: :any_skip_relocation, arm64_monterey: "'$$SHA256_DARWIN_ARM64'"/' \
			-e 's/sha256 cellar: :any_skip_relocation, sonoma: "[^"]*"/sha256 cellar: :any_skip_relocation, sonoma: "'$$SHA256_DARWIN_AMD64'"/' \
			-e 's/sha256 cellar: :any_skip_relocation, ventura: "[^"]*"/sha256 cellar: :any_skip_relocation, ventura: "'$$SHA256_DARWIN_AMD64'"/' \
			-e 's/sha256 cellar: :any_skip_relocation, monterey: "[^"]*"/sha256 cellar: :any_skip_relocation, monterey: "'$$SHA256_DARWIN_AMD64'"/' \
			-e 's/sha256 cellar: :any_skip_relocation, x86_64_linux: "[^"]*"/sha256 cellar: :any_skip_relocation, x86_64_linux: "'$$SHA256_LINUX_AMD64'"/' \
			-e 's/sha256 cellar: :any_skip_relocation, aarch64_linux: "[^"]*"/sha256 cellar: :any_skip_relocation, aarch64_linux: "'$$SHA256_LINUX_ARM64'"/' \
			$(FORMULA_FILE) && \
		echo "Creating pull request..." && \
		git config user.name "GitHub Actions" && \
		git config user.email "github-actions@github.com" && \
		git add $(FORMULA_FILE) && \
		git commit -m "chore: update mdctl to $(VERSION)" && \
		git push -u origin $(BRANCH_NAME) && \
		curl -X POST \
			-H "Accept: application/vnd.github.v3+json" \
			-H "Authorization: token $(GH_PAT)" \
			https://api.github.com/repos/samzong/$(HOMEBREW_TAP_REPO)/pulls \
			-d '{"title":"chore: update mdctl to $(VERSION)","body":"Update mdctl to $(VERSION)\n\n- Version: $(CLEAN_VERSION)\n- SHA256 Darwin AMD64: '$$SHA256_DARWIN_AMD64'\n- SHA256 Darwin ARM64: '$$SHA256_DARWIN_ARM64'\n- SHA256 Linux AMD64: '$$SHA256_LINUX_AMD64'\n- SHA256 Linux ARM64: '$$SHA256_LINUX_ARM64'","head":"$(BRANCH_NAME)","base":"main"}'
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
