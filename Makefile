.PHONY: help build test test-short test-coverage lint clean install run fmt vet mod mod-tidy mod-verify release snapshot check pre-release all setup bench build-all generate

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

setup: ## Install development dependencies
	@echo "Installing development tools..."
	$(GO) install tool

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

test-short: ## Run short tests only (skip integration tests)
	@echo "Running short tests..."
	$(GO) test -v -short -race ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

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

mod-tidy: ## Verify go.mod and go.sum are tidy
	@echo "Verifying modules are tidy..."
	$(GO) mod tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod/go.sum not tidy" && exit 1)

mod-verify: ## Verify module dependencies
	@echo "Verifying module dependencies..."
	$(GO) mod verify

generate: ## Run go generate
	@echo "Running go generate..."
	$(GO) generate ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-*
	@rm -f coverage.out coverage.html

install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) $(MAIN_PATH)

run: ## Run the application
	$(GO) run $(MAIN_PATH)

pre-release: mod-tidy check ## Run all checks before release
	@echo "Running pre-release checks..."
	$(GO) mod verify
	@echo "Pre-release checks passed"

release: pre-release ## Build release binaries with goreleaser
	@echo "Building release with goreleaser..."
	$(GORELEASER) release --clean

snapshot: ## Build snapshot release with goreleaser (no publish)
	@echo "Building snapshot with goreleaser..."
	$(GORELEASER) release --snapshot --clean

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

all: clean mod check build ## Run clean, mod, check, and build

.DEFAULT_GOAL := help
