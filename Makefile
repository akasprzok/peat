.PHONY: help build test lint clean install run fmt vet mod release snapshot

# Variables
BINARY_NAME=peat
BUILD_DIR=dist
MAIN_PATH=.
GO=go
GORELEASER=goreleaser

# Build information
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run golangci-lint
	@echo "Running linter..."
	$(GO) tool golangci-lint run

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GO) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

mod: ## Download and tidy Go modules
	@echo "Tidying modules..."
	$(GO) mod tidy
	$(GO) mod download

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html

install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) $(MAIN_PATH)

run: ## Run the application
	$(GO) run $(MAIN_PATH)

release: ## Build release binaries with goreleaser
	@echo "Building release with goreleaser..."
	$(GORELEASER) release --clean

snapshot: ## Build snapshot release with goreleaser (no publish)
	@echo "Building snapshot with goreleaser..."
	$(GORELEASER) release --snapshot --clean

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

all: clean mod check build ## Run clean, mod, check, and build

.DEFAULT_GOAL := help

