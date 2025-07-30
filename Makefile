.PHONY: build clean test lint install dev help check-deps release release-dry release-patch release-minor release-major api-build api-run api-dev swagger-gen swagger-serve
.DEFAULT_GOAL := help

# Application info
APP_NAME := psi-map
MAIN_PACKAGE := ./cmd/cli/main.go
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

# Docker vars
DOCKER_TAG ?= $(VERSION)
DOCKER_IMAGE ?= psi-map

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

# API specific targets
#
# Generate Swagger documentation
swagger-gen:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go -o docs/
	@echo "Swagger docs generated in docs/"

# Serve Swagger documentation locally
swagger-serve:
	@echo "Serving Swagger docs at http://localhost:8081"
	@cd docs && python3 -m http.server 8081

# Build API binary
api-build: swagger-gen
	@echo "Building API binary..."
	@go build -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)" \
		-o bin/psi-map-api cmd/api/main.go
	@echo "API binary built: bin/psi-map-api"

# Run API server
api-run: api-build
	@echo "Starting PSI-Map API server..."
	@./bin/psi-map-api

# Development mode with auto-reload (requires air: go install github.com/cosmtrek/air@latest)
api-dev: swagger-gen
	@echo "Starting API in development mode..."
	@air -c .air-api.toml

# Test API endpoints (requires httpie: pip install httpie)
api-test:
	@echo "Testing API endpoints..."
	@echo "Health check:"
	@http GET localhost:8080/api/v1/health
	@echo "\nVersion info:"
	@http GET localhost:8080/api/v1/version
	@echo "\nSync analysis (replace with real sitemap):"
	@http POST localhost:8080/api/v1/analyze sitemap="https://example.com/sitemap.xml" async:=false max_workers:=2
	@echo "\nAsync analysis:"
	@http POST localhost:8080/api/v1/analyze sitemap="https://example.com/sitemap.xml" async:=true max_workers:=2

# Docker targets for API
docker-api-build:
	@echo "Building API Docker image..."
	@docker build -f Dockerfile.api -t psi-map-api:$(VERSION) .

docker-api-run:
	@echo "Running API in Docker..."
	@docker run -p 8080:8080 -e PORT=8080 psi-map-api:$(VERSION)
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


release-dry: ## Run a dry-run release locally with goreleaser
	@echo "$(BLUE)Running goreleaser dry run...$(NC)"
	goreleaser release --snapshot --clean
	@echo "$(GREEN)✓ Dry run release completed$(NC)"

release-patch: ## Tag and push a patch release (triggers GitHub release)
	@echo "$(BLUE)Tagging patch release...$(NC)"
	./scripts/release.sh patch
	@echo "$(GREEN)✓ Patch release tagged and pushed$(NC)"

release-minor: ## Tag and push a minor release (triggers GitHub release)
	@echo "$(BLUE)Tagging minor release...$(NC)"
	./scripts/release.sh minor
	@echo "$(GREEN)✓ Minor release tagged and pushed$(NC)"

release-major: ## Tag and push a major release (triggers GitHub release)
	@echo "$(BLUE)Tagging major release...$(NC)"
	./scripts/release.sh major
	@echo "$(GREEN)✓ Major release tagged and pushed$(NC)"

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

# Build Docker image with both CLI and API
docker-build:
	@echo "Building Docker image with both CLI and API..."
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run API in Docker
docker-api:
	@echo "Starting PSI-Map API in Docker..."
	@docker run -d \
		--name psi-map-api \
		-p 8080:8080 \
		-e PORT=8080 \
		--entrypoint /app/psi-map-api \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "API running at http://localhost:8080"
	@echo "Swagger docs at http://localhost:8080/swagger/"

# Run CLI in Docker
docker-cli:
	@echo "Running PSI-Map CLI in Docker..."
	@docker run --rm \
		-v $(PWD)/output:/app/output \
		-v $(PWD)/cache:/app/cache \
		--entrypoint /app/psi-map \
		$(DOCKER_IMAGE):$(DOCKER_TAG) $(ARGS)

# Docker Compose commands
compose-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d psi-map-api
	@echo "API available at http://localhost:8080"

compose-dev:
	@echo "Starting in development mode..."
	@docker-compose -f docker-compose.yml -f docker-compose.override.yml up

compose-prod:
	@echo "Starting in production mode..."
	@docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

compose-down:
	@docker-compose down

compose-logs:
	@docker-compose logs -f psi-map-api

# CLI via Docker Compose
compose-cli:
	@echo "Running CLI via Docker Compose..."
	@docker-compose run --rm psi-map-cli $(ARGS)

# Examples
docker-examples:
	@echo "=== Docker Usage Examples ==="
	@echo ""
	@echo "1. Build image:"
	@echo "   make docker-build"
	@echo ""
	@echo "2. Run API:"
	@echo "   make docker-api"
	@echo ""
	@echo "3. Run CLI analysis:"
	@echo "   make docker-cli ARGS='analyze sitemap.xml'"
	@echo ""
	@echo "4. Run with Docker Compose:"
	@echo "   make compose-up"
	@echo ""
	@echo "5. CLI with Compose:"
	@echo "   make compose-cli ARGS='analyze --output html sitemap.xml'"
	@echo ""
	@echo "6. Development mode:"
	@echo "   make compose-dev"

# Clean up Docker resources
docker-clean:
	@echo "Cleaning up Docker resources..."
	-@docker stop psi-map-api 2>/dev/null || true
	-@docker rm psi-map-api 2>/dev/null || true
	@docker-compose down 2>/dev/null || true
	@echo "Cleanup complete"

# Push to registry (customize for your registry)
docker-push: docker-build
	@echo "Pushing to registry..."
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker push $(DOCKER_IMAGE):latest

## CI/CD helpers

ci-setup: deps install-tools ## Set up CI environment (dependencies and tools)
	@echo "$(GREEN)✓ CI environment ready$(NC)"

ci-test: deps test-coverage lint ## Run CI tests
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
