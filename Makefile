# Professional Azure Container Deployment CLI Tool
# Build and development targets

# Variables
BIN_NAME=azctl
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT) -s -w"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=$(BIN_NAME)
BINARY_UNIX=$(BIN_NAME)_unix

# Directories
BIN_DIR=bin
DIST_DIR=dist
COVERAGE_DIR=coverage
DOCS_DIR=docs

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
WHITE=\033[0;37m
NC=\033[0m # No Color

.PHONY: help build test lint clean release install-tools check-deps

# Default target
all: clean build test lint

# Help target
help: ## Show this help message
	@echo "$(CYAN)Professional Azure Container Deployment CLI Tool$(NC)"
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
build: ## Build the binary
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/azctl
	@echo "$(GREEN)✓ Built $(BINARY_NAME) successfully$(NC)"

test: ## Run tests with coverage
	@echo "$(BLUE)Running tests...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)✓ Tests completed$(NC)"
	@echo "$(CYAN)Coverage report: $(COVERAGE_DIR)/coverage.html$(NC)"

test-unit: ## Run unit tests only
	@echo "$(BLUE)Running unit tests...$(NC)"
	$(GOTEST) -v -race ./internal/...

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	$(GOTEST) -v -race -tags=integration ./...

lint: ## Run linters
	@echo "$(BLUE)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "$(YELLOW)golangci-lint not found, installing...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run --timeout=5m; \
	fi
	@echo "$(GREEN)✓ Linting completed$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GOCMD) vet ./...
	@echo "$(GREEN)✓ Go vet completed$(NC)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOCMD) fmt ./...
	@echo "$(GREEN)✓ Code formatting completed$(NC)"

# Build targets
build-all: ## Build for all platforms
	@echo "$(BLUE)Building for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_linux_amd64 ./cmd/azctl
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_linux_arm64 ./cmd/azctl
	
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_darwin_amd64 ./cmd/azctl
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_darwin_arm64 ./cmd/azctl
	
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_windows_amd64.exe ./cmd/azctl
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_windows_arm64.exe ./cmd/azctl
	
	@echo "$(GREEN)✓ Built for all platforms$(NC)"

release: build-all ## Build release binaries
	@echo "$(BLUE)Creating release...$(NC)"
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --rm-dist; \
	else \
		echo "$(YELLOW)goreleaser not found, using build-all$(NC)"; \
	fi
	@echo "$(GREEN)✓ Release created$(NC)"

# Cleanup targets
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -rf $(BIN_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@echo "$(GREEN)✓ Cleanup completed$(NC)"

# Dependency management
deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

check-deps: ## Check for outdated dependencies
	@echo "$(BLUE)Checking for outdated dependencies...$(NC)"
	$(GOCMD) list -u -m all
	@echo "$(GREEN)✓ Dependency check completed$(NC)"

# Installation targets
install: build ## Install binary to system
	@echo "$(BLUE)Installing $(BINARY_NAME)...$(NC)"
	@if [ "$(shell uname)" = "Darwin" ]; then \
		sudo cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
	elif [ "$(shell uname)" = "Linux" ]; then \
		sudo cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
	else \
		echo "$(YELLOW)Installation not supported on this platform$(NC)"; \
	fi
	@echo "$(GREEN)✓ $(BINARY_NAME) installed$(NC)"

install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/goreleaser/goreleaser@latest
	$(GOGET) golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)✓ Development tools installed$(NC)"

# Documentation targets
docs: ## Generate documentation
	@echo "$(BLUE)Generating documentation...$(NC)"
	@mkdir -p $(DOCS_DIR)
	@echo "# $(BINARY_NAME) Documentation" > $(DOCS_DIR)/README.md
	@echo "Generated on $(BUILD_TIME)" >> $(DOCS_DIR)/README.md
	@echo "$(GREEN)✓ Documentation generated$(NC)"

# Security targets
security: ## Run security checks
	@echo "$(BLUE)Running security checks...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec not found, installing...$(NC)"; \
		$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi
	@echo "$(GREEN)✓ Security checks completed$(NC)"

# Performance targets
bench: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...
	@echo "$(GREEN)✓ Benchmarks completed$(NC)"

# CI/CD targets
ci: ## Run CI pipeline
	@echo "$(BLUE)Running CI pipeline...$(NC)"
	$(MAKE) deps
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) security
	$(MAKE) build
	@echo "$(GREEN)✓ CI pipeline completed$(NC)"

# Development workflow
dev: ## Development workflow
	@echo "$(BLUE)Starting development workflow...$(NC)"
	$(MAKE) deps
	$(MAKE) build
	$(MAKE) test
	@echo "$(GREEN)✓ Development workflow completed$(NC)"

# Version information
version: ## Show version information
	@echo "$(CYAN)Version: $(VERSION)$(NC)"
	@echo "$(CYAN)Build Time: $(BUILD_TIME)$(NC)"
	@echo "$(CYAN)Git Commit: $(GIT_COMMIT)$(NC)"

# Debug targets
debug: ## Show debug information
	@echo "$(BLUE)Debug information:$(NC)"
	@echo "  Go version: $(shell go version)"
	@echo "  OS: $(shell uname -s)"
	@echo "  Architecture: $(shell uname -m)"
	@echo "  Working directory: $(shell pwd)"
	@echo "  Go modules: $(shell go env GOMOD)"


