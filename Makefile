.PHONY: all build clean run test lint fmt help pre-commit-install pre-commit-run setup release

# Binary name
BINARY_NAME=jellywolproxy

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Use the Go version from go.mod
GO_VERSION=$(shell grep -E '^go [0-9]+\.[0-9]+(\.[0-9]+)?$$' go.mod | cut -d ' ' -f 2)

# Build variables
BUILD_DIR=build
VERSION?=0.0.12
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: clean build ## Clean and build the project

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build ${LDFLAGS} -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/...

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

run: build ## Run the application
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

release: ## Create a new release (usage: make release VERSION=v1.0.0)
	@if [ "$(VERSION)" = "" ]; then \
		echo "Error: VERSION is required. Use: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@if [ -n "`git status --porcelain`" ]; then \
		echo "Error: working directory is not clean. Please commit or stash changes first."; \
		exit 1; \
	fi
	@echo "Creating new release $(VERSION)..."
	@git tag -a $(VERSION) -m "$(VERSION)"
	@git push origin $(VERSION)
	@echo "Release $(VERSION) created and pushed. GitHub Actions will build and publish the release."

lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint is not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

update-deps: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

check-go-version: ## Check if correct Go version is installed
	@echo "Checking Go version..."
	@if command -v go >/dev/null; then \
		current_version=$$(go version | cut -d " " -f 3 | sed 's/go//'); \
		if [ "$$current_version" != "$(GO_VERSION)" ]; then \
			echo "Warning: Current Go version ($$current_version) does not match go.mod version ($(GO_VERSION))"; \
		else \
			echo "Go version matches go.mod ($(GO_VERSION))"; \
		fi \
	else \
		echo "Go is not installed"; \
		exit 1; \
	fi

setup: ## Install all development dependencies
	@echo "Installing development dependencies..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@go install github.com/go-critic/go-critic/cmd/gocritic@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go mod tidy
	@make pre-commit-install

pre-commit-install: ## Install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@if command -v pre-commit >/dev/null; then \
		pre-commit install; \
	else \
		echo "pre-commit is not installed. Please install it first with 'brew install pre-commit'"; \
		exit 1; \
	fi

pre-commit-run: ## Run pre-commit hooks manually
	@echo "Running pre-commit hooks..."
	@if command -v pre-commit >/dev/null; then \
		pre-commit run --all-files; \
	else \
		echo "pre-commit is not installed. Please install it first with 'brew install pre-commit'"; \
		exit 1; \
	fi

# Default target
.DEFAULT_GOAL := help
