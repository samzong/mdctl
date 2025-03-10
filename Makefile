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

# Adjust architecture definitions to match goreleaser output
SUPPORTED_ARCHS = Darwin_x86_64 Darwin_arm64 Linux_x86_64 Linux_arm64

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
	@echo "==> Starting Homebrew formula update process..."
	@if [ -z "$(GH_PAT)" ]; then \
		echo "❌ Error: GH_PAT environment variable is required"; \
		exit 1; \
	fi

	@echo "==> Current version information:"
	@echo "    - VERSION: $(VERSION)"
	@echo "    - CLEAN_VERSION: $(CLEAN_VERSION)"

	@echo "==> Preparing working directory..."
	@rm -rf tmp && mkdir -p tmp
	
	@echo "==> Cloning Homebrew tap repository..."
	@cd tmp && git clone https://$(GH_PAT)@github.com/samzong/$(HOMEBREW_TAP_REPO).git
	@cd tmp/$(HOMEBREW_TAP_REPO) && echo "    - Creating new branch: $(BRANCH_NAME)" && git checkout -b $(BRANCH_NAME)

	@echo "==> Processing architectures and calculating checksums..."
	@cd tmp/$(HOMEBREW_TAP_REPO) && \
	for arch in $(SUPPORTED_ARCHS); do \
		echo "    - Processing $$arch..."; \
		if [ "$(DRY_RUN)" = "1" ]; then \
			echo "      [DRY_RUN] Would download: https://github.com/samzong/mdctl/releases/download/v$(CLEAN_VERSION)/mdctl_$${arch}.tar.gz"; \
			case "$$arch" in \
				Darwin_x86_64) DARWIN_AMD64_SHA="fake_sha_amd64" ;; \
				Darwin_arm64) DARWIN_ARM64_SHA="fake_sha_arm64" ;; \
				Linux_x86_64) LINUX_AMD64_SHA="fake_sha_linux_amd64" ;; \
				Linux_arm64) LINUX_ARM64_SHA="fake_sha_linux_arm64" ;; \
			esac; \
		else \
			echo "      - Downloading release archive..."; \
			curl -L -sSfO "https://github.com/samzong/mdctl/releases/download/v$(CLEAN_VERSION)/mdctl_$${arch}.tar.gz" || { echo "❌ Failed to download $$arch archive"; exit 1; }; \
			echo "      - Calculating SHA256..."; \
			sha=$$(shasum -a 256 "mdctl_$${arch}.tar.gz" | cut -d' ' -f1); \
			case "$$arch" in \
				Darwin_x86_64) DARWIN_AMD64_SHA="$$sha"; echo "      ✓ Darwin AMD64 SHA: $$sha" ;; \
				Darwin_arm64) DARWIN_ARM64_SHA="$$sha"; echo "      ✓ Darwin ARM64 SHA: $$sha" ;; \
				Linux_x86_64) LINUX_AMD64_SHA="$$sha"; echo "      ✓ Linux AMD64 SHA: $$sha" ;; \
				Linux_arm64) LINUX_ARM64_SHA="$$sha"; echo "      ✓ Linux ARM64 SHA: $$sha" ;; \
			esac; \
		fi; \
	done; \
	\
	if [ "$(DRY_RUN)" = "1" ]; then \
		echo "==> [DRY_RUN] Would update formula with:"; \
		echo "    - Darwin AMD64 SHA: $$DARWIN_AMD64_SHA"; \
		echo "    - Darwin ARM64 SHA: $$DARWIN_ARM64_SHA"; \
		echo "    - Linux AMD64 SHA: $$LINUX_AMD64_SHA"; \
		echo "    - Linux ARM64 SHA: $$LINUX_ARM64_SHA"; \
		echo "    - Would commit and push changes"; \
		echo "    - Would create PR"; \
	else \
		echo "==> Updating formula file..."; \
		echo "    - Updating version to $(CLEAN_VERSION)"; \
		sed -i '' -e 's|version ".*"|version "$(CLEAN_VERSION)"|' $(FORMULA_FILE); \
		\
		echo "    - Updating URLs and checksums"; \
		sed -i '' \
			-e '/on_macos/,/end/ { \
				/if Hardware::CPU.arm?/,/else/ { \
					s|url ".*"|url "https://github.com/samzong/mdctl/releases/download/v#{version}/mdctl_Darwin_arm64.tar.gz"|; \
					s|sha256 ".*"|sha256 "'"$$DARWIN_ARM64_SHA"'"|; \
				}; \
				/else/,/end/ { \
					s|url ".*"|url "https://github.com/samzong/mdctl/releases/download/v#{version}/mdctl_Darwin_x86_64.tar.gz"|; \
					s|sha256 ".*"|sha256 "'"$$DARWIN_AMD64_SHA"'"|; \
				}; \
			}' \
			-e '/on_linux/,/end/ { \
				/if Hardware::CPU.arm?/,/else/ { \
					s|url ".*"|url "https://github.com/samzong/mdctl/releases/download/v#{version}/mdctl_Linux_arm64.tar.gz"|; \
					s|sha256 ".*"|sha256 "'"$$LINUX_ARM64_SHA"'"|; \
				}; \
				/else/,/end/ { \
					s|url ".*"|url "https://github.com/samzong/mdctl/releases/download/v#{version}/mdctl_Linux_x86_64.tar.gz"|; \
					s|sha256 ".*"|sha256 "'"$$LINUX_AMD64_SHA"'"|; \
				}; \
			}' $(FORMULA_FILE); \
		\
		echo "    - Checking for changes..."; \
		if ! git diff --quiet $(FORMULA_FILE); then \
			echo "==> Changes detected, creating pull request..."; \
			echo "    - Adding changes to git"; \
			git add $(FORMULA_FILE); \
			echo "    - Committing changes"; \
			git commit -m "chore: bump to $(VERSION)"; \
			echo "    - Pushing to remote"; \
			git push -u origin $(BRANCH_NAME); \
			echo "    - Preparing pull request data"; \
			pr_data=$$(jq -n \
				--arg title "chore: update mdctl to $(VERSION)" \
				--arg body "Auto-generated PR\nSHAs:\n- Darwin(amd64): $$DARWIN_AMD64_SHA\n- Darwin(arm64): $$DARWIN_ARM64_SHA" \
				--arg head "$(BRANCH_NAME)" \
				--arg base "main" \
				'{title: $$title, body: $$body, head: $$head, base: $$base}'); \
			echo "    - Creating pull request"; \
			curl -X POST \
				-H "Authorization: token $(GH_PAT)" \
				-H "Content-Type: application/json" \
				https://api.github.com/repos/samzong/$(HOMEBREW_TAP_REPO)/pulls \
				-d "$$pr_data"; \
			echo "✅ Pull request created successfully"; \
		else \
			echo "❌ No changes detected in formula file"; \
			exit 1; \
		fi; \
	fi

	@echo "==> Cleaning up temporary files..."
	@rm -rf tmp
	@echo "✅ Homebrew formula update process completed"

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
