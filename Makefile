.PHONY: build clean test lint install dev help release check-deps
.DEFAULT_GOAL := help

# Application info
APP_NAME := psi-map
MAIN_PACKAGE := ./main.go
DIST_DIR := dist
COVERAGE_DIR := coverage

GO_VERSION := 1.24
export GO_VERSION

# Version info - can be overridden in CI
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME) -s -w"
BUILD_FLAGS := -trimpath $(LDFLAGS)

# Cross-compilation targets
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

## Development commands

dev: ## Build and run in development mode
	@echo "$(BLUE)Building $(APP_NAME) for development...$(NC)"
	go run $(MAIN_PACKAGE) ${ARGS}

build: clean ## Build binary for current platform
	@echo "$(BLUE)Building $(APP_NAME) v$(VERSION)...$(NC)"
	@mkdir -p $(DIST_DIR)
	go build $(BUILD_FLAGS) -o $(DIST_DIR)/$(APP_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)✓ Built $(DIST_DIR)/$(APP_NAME)$(NC)"

install: ## Install binary to $GOPATH/bin
	@echo "$(BLUE)Installing $(APP_NAME)...$(NC)"
	go install $(BUILD_FLAGS) $(MAIN_PACKAGE)
	@echo "$(GREEN)✓ Installed $(APP_NAME) to $(shell go env GOPATH)/bin$(NC)"

## Testing and quality

test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	go test -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)✓ Coverage report: $(COVERAGE_DIR)/coverage.html$(NC)"

lint: check-deps ## Run linting
	@echo "$(BLUE)Running linting...$(NC)"
	golangci-lint run ./...
	@echo "$(GREEN)✓ Linting passed$(NC)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)✓ go vet passed$(NC)"

mod-tidy: ## Tidy go modules
	@echo "$(BLUE)Tidying go modules...$(NC)"
	go mod tidy
	@echo "$(GREEN)✓ Go modules tidied$(NC)"

## Release and distribution

release: clean test lint build-all ## Create a full release (test, lint, build all platforms)
	@echo "$(GREEN)✓ Release $(VERSION) ready in $(DIST_DIR)/$(NC)"
	@ls -la $(DIST_DIR)/

build-all: clean ## Build binaries for all platforms
	@echo "$(BLUE)Building $(APP_NAME) v$(VERSION) for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@$(foreach platform,$(PLATFORMS), \
		echo "Building for $(platform)..."; \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		go build $(BUILD_FLAGS) \
			-o $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$(subst /,-,$(platform))$(if $(findstring windows,$(platform)),.exe,) \
			$(MAIN_PACKAGE) && \
		echo "$(GREEN)✓ Built for $(platform)$(NC)" || \
		(echo "$(RED)✗ Failed to build for $(platform)$(NC)" && exit 1); \
	)

checksums: ## Generate checksums for release binaries
	@echo "$(BLUE)Generating checksums...$(NC)"
	@cd $(DIST_DIR) && \
	for file in $(APP_NAME)-*; do \
		if [ -f "$$file" ]; then \
			sha256sum "$$file" >> checksums.txt; \
		fi; \
	done
	@echo "$(GREEN)✓ Checksums generated: $(DIST_DIR)/checksums.txt$(NC)"

## Utility commands

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(DIST_DIR) $(COVERAGE_DIR)
	@go clean
	@echo "$(GREEN)✓ Cleaned$(NC)"

deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)✓ Development tools installed$(NC)"

check-deps: ## Check if required tools are installed
	@echo "$(BLUE)Checking dependencies...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 && echo "$(GREEN)✓ golangci-lint found$(NC)" || echo "$(RED)✗ golangci-lint not found$(NC)"

info: ## Show build info
	@echo "$(BLUE)Build Information:$(NC)"
	@echo "  App Name:    $(APP_NAME)"
	@echo "  Version:     $(VERSION)"
	@echo "  Commit:      $(COMMIT)"
	@echo "  Build Time:  $(BUILD_TIME)"
	@echo "  Go Version:  $(shell go version)"
	@echo "  Platforms:   $(PLATFORMS)"

docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .
	@echo "$(GREEN)✓ Docker image built: $(APP_NAME):$(VERSION)$(NC)"

docker-build: ## Build and push Docker image for CI
	@echo "$(BLUE)Building and pushing Docker image...$(NC)"
	@docker buildx create --use
	@docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t ghcr.io/$(REPO):$(VERSION) \
		-t ghcr.io/$(REPO):latest \
		--push \
		--cache-from=type=gha \
		--cache-to=type=gha,mode=max \
		.
	@echo "$(GREEN)✓ Docker image pushed: ghcr.io/$(REPO):$(VERSION)$(NC)"

## CI/CD helpers

ci-setup: deps install-tools ## Set up CI environment (dependencies and tools)
	@echo "$(GREEN)✓ CI environment ready$(NC)"

ci-test: deps test lint ## Run CI tests
	@echo "$(GREEN)✓ CI tests completed$(NC)"

ci-build: clean build-all checksums ## Build for CI
	@echo "$(GREEN)✓ CI build completed$(NC)"

ci-release: ci-test ci-build ## Full CI release process
	@echo "$(GREEN)✓ CI release completed$(NC)"

## Help

help: ## Show this help message
	@echo "$(BLUE)$(APP_NAME) - PageSpeed Insights CLI Tool$(NC)"
	@echo ""
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make <target>"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' '/^dev:|^build:|^install:/ {print "  " $$1 ": " $$2}' | sort
	@echo ""
	@echo "$(YELLOW)Testing & Quality:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' '/^test:|^test-coverage:|^lint:|^fmt:|^vet:|^mod-tidy:/ {print "  " $$1 ": " $$2}' | sort
	@echo ""
	@echo "$(YELLOW)Release & Distribution:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' '/^release:|^build-all:|^checksums:/ {print "  " $$1 ": " $$2}' | sort
	@echo ""
	@echo "$(YELLOW)Utilities:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' '/^clean:|^deps:|^check-deps:|^info:|^docker-build:/ {print "  " $$1 ": " $$2}' | sort
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make build          # Build for current platform"
	@echo "  make test           # Run tests"
	@echo "  make release        # Full release build"
	@echo "  make ci-release     # CI release process"
