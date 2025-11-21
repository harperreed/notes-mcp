# ABOUTME: Makefile for building, testing, and managing the Apple Notes MCP server
# ABOUTME: Provides convenient targets for development, testing, and deployment

.PHONY: help build test test-integration test-all lint clean install run dev format check pre-commit

# Binary name
BINARY_NAME=notes-mcp
INSTALL_PATH=/usr/local/bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-v

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

build-release: ## Build optimized release binary
	@echo "Building release binary..."
	$(GOBUILD) -v -ldflags "-s -w -X main.version=$$(git describe --tags --always --dirty)" -o $(BINARY_NAME) .
	@echo "Release build complete: ./$(BINARY_NAME)"

test: ## Run unit tests
	@echo "Running unit tests..."
	$(GOTEST) -v -race -cover ./...

test-integration: ## Run integration tests (requires Apple Notes)
	@echo "Running integration tests..."
	$(GOTEST) -v -race -tags=integration ./...

test-all: ## Run all tests (unit + integration)
	@echo "Running all tests..."
	$(GOTEST) -v -race -cover -tags=integration ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run golangci-lint
	@echo "Running linter..."
	$(GOLINT) run --timeout=5m

lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running linter with auto-fix..."
	$(GOLINT) run --timeout=5m --fix

format: ## Format code with gofmt and goimports
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

check: format lint test ## Run format, lint, and test

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

install: build ## Install binary to system
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	install -m 755 $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_PATH)/$(BINARY_NAME)"

uninstall: ## Uninstall binary from system
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled"

run: build ## Build and run the MCP server
	@echo "Starting MCP server..."
	./$(BINARY_NAME) mcp

dev: ## Run with live reload (requires entr)
	@echo "Starting development mode with auto-reload..."
	@command -v entr >/dev/null 2>&1 || { echo "entr not installed. Install with: brew install entr"; exit 1; }
	@find . -name '*.go' | entr -r make run

pre-commit: ## Run pre-commit hooks manually
	@echo "Running pre-commit hooks..."
	pre-commit run --all-files

tidy: ## Tidy go.mod and go.sum
	@echo "Tidying go modules..."
	$(GOMOD) tidy
	$(GOMOD) verify

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

upgrade-deps: ## Upgrade all dependencies
	@echo "Upgrading dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# CLI command targets for convenience
create: build ## Create a test note (example)
	./$(BINARY_NAME) create "Test Note" "This is a test note created via Makefile" --tags=test,makefile

search: build ## Search for notes (example: make search QUERY="test")
	./$(BINARY_NAME) search "$(QUERY)"

folders: build ## List all folders
	./$(BINARY_NAME) folders

# Docker targets (if you want to add Docker support later)
docker-build: ## Build Docker image
	@echo "Docker support not yet implemented"

docker-run: ## Run Docker container
	@echo "Docker support not yet implemented"

# Version info
version: ## Show version information
	@echo "Binary: $(BINARY_NAME)"
	@echo "Go version: $$($(GOCMD) version)"
	@echo "Git commit: $$(git rev-parse --short HEAD)"
	@echo "Git tag: $$(git describe --tags --always --dirty)"

.DEFAULT_GOAL := help
