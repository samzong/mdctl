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

# 调整架构定义匹配goreleaser输出
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
	@if [ -z "$(GH_PAT)" ]; then \
		echo "Error: GH_PAT required"; \
		exit 1; \
	fi
	@echo "Updating formula for v$(CLEAN_VERSION)"
	@rm -rf tmp && mkdir -p tmp
	@cd tmp && \
		git clone https://$(GH_PAT)@github.com/samzong/$(HOMEBREW_TAP_REPO).git && \
		cd $(HOMEBREW_TAP_REPO) && \
		git checkout -b $(BRANCH_NAME)
	@(cd tmp/$(HOMEBREW_TAP_REPO) && \
		for arch in $(SUPPORTED_ARCHS); do \
			echo "Processing $$arch..." && \
			if [ "$(DRY_RUN)" = "1" ]; then \
				echo "[DRY_RUN] Would download: https://github.com/samzong/mdctl/releases/download/v$(CLEAN_VERSION)/mdctl_$${arch}.tar.gz" && \
				case "$$arch" in \
					Darwin_x86_64) DARWIN_AMD64_SHA="fake_sha_amd64" ;; \
					Darwin_arm64) DARWIN_ARM64_SHA="fake_sha_arm64" ;; \
					Linux_x86_64) LINUX_AMD64_SHA="fake_sha_linux_amd64" ;; \
					Linux_arm64) LINUX_ARM64_SHA="fake_sha_linux_arm64" ;; \
				esac; \
			else \
				curl -L -sSfO "https://github.com/samzong/mdctl/releases/download/v$(CLEAN_VERSION)/mdctl_$${arch}.tar.gz" || exit 1 && \
				sha=$$(shasum -a 256 "mdctl_$${arch}.tar.gz" | cut -d' ' -f1) && \
				case "$$arch" in \
					Darwin_x86_64) DARWIN_AMD64_SHA="$$sha" ;; \
					Darwin_arm64) DARWIN_ARM64_SHA="$$sha" ;; \
					Linux_x86_64) LINUX_AMD64_SHA="$$sha" ;; \
					Linux_arm64) LINUX_ARM64_SHA="$$sha" ;; \
				esac; \
			fi \
		done && \
		if [ "$(DRY_RUN)" = "1" ]; then \
			echo "[DRY_RUN] Would update formula with:" && \
			echo "  - Darwin AMD64 SHA: $$DARWIN_AMD64_SHA" && \
			echo "  - Darwin ARM64 SHA: $$DARWIN_ARM64_SHA" && \
			echo "  - Linux AMD64 SHA: $$LINUX_AMD64_SHA" && \
			echo "  - Linux ARM64 SHA: $$LINUX_ARM64_SHA" && \
			echo "[DRY_RUN] Would commit and push changes" && \
			echo "[DRY_RUN] Would create PR"; \
		else \
			sed -i '' \
				-e 's|url ".*/mdctl_Darwin_arm64\.tar\.gz"|url "https://github.com/samzong/mdctl/releases/download/v#{version}/mdctl_Darwin_arm64.tar.gz"|' \
				-e "s|sha256 \".*\" # Darwin_arm64|sha256 \"$$DARWIN_ARM64_SHA\" # Darwin_arm64|" \
				-e 's|url ".*/mdctl_Darwin_x86_64\.tar\.gz"|url "https://github.com/samzong/mdctl/releases/download/v#{version}/mdctl_Darwin_x86_64.tar.gz"|' \
				-e "s|sha256 \".*\" # Darwin_x86_64|sha256 \"$$DARWIN_AMD64_SHA\" # Darwin_x86_64|" \
				-e 's|url ".*/mdctl_Linux_arm64\.tar\.gz"|url "https://github.com/samzong/mdctl/releases/download/v#{version}/mdctl_Linux_arm64.tar.gz"|' \
				-e "s|sha256 \".*\" # Linux_arm64|sha256 \"$$LINUX_ARM64_SHA\" # Linux_arm64|" \
				-e 's|url ".*/mdctl_Linux_x86_64\.tar\.gz"|url "https://github.com/samzong/mdctl/releases/download/v#{version}/mdctl_Linux_x86_64.tar.gz"|' \
				-e "s|sha256 \".*\" # Linux_x86_64|sha256 \"$$LINUX_AMD64_SHA\" # Linux_x86_64|" \
				$(FORMULA_FILE) && \
			if ! git diff --quiet $(FORMULA_FILE); then \
				git add $(FORMULA_FILE) && \
				git commit -m "chore: bump to $(VERSION)" && \
				git push -u origin $(BRANCH_NAME) && \
				pr_data=$$(jq -n \
					--arg title "chore: update mdctl to $(VERSION)" \
					--arg body "Auto-generated PR\nSHAs:\n- Darwin(amd64): $$DARWIN_AMD64_SHA\n- Darwin(arm64): $$DARWIN_ARM64_SHA" \
					--arg head "$(BRANCH_NAME)" \
					--arg base "main" \
					'{title: $$title, body: $$body, head: $$head, base: $$base}') && \
				curl -X POST \
					-H "Authorization: token $(GH_PAT)" \
					-H "Content-Type: application/json" \
					https://api.github.com/repos/samzong/$(HOMEBREW_TAP_REPO)/pulls \
					-d "$$pr_data"; \
			else \
				echo "No changes detected" && \
				exit 1; \
			fi \
		fi \
	)
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
