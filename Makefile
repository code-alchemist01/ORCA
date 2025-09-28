# Orca Orchestrator Makefile

# Variables
BINARY_NAME=orca
ORCHESTRATOR_BINARY=orchestrator
CLI_BINARY=orcacli
BUILD_DIR=bin
GO_VERSION=1.21

# Default target
.PHONY: all
all: clean build

# Clean build directory
.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
	@mkdir $(BUILD_DIR)

# Build all binaries
.PHONY: build
build: build-orchestrator build-cli

# Build orchestrator
.PHONY: build-orchestrator
build-orchestrator:
	@echo "Building orchestrator..."
	go build -o $(BUILD_DIR)/$(ORCHESTRATOR_BINARY).exe ./cmd/orchestrator

# Build CLI
.PHONY: build-cli
build-cli:
	@echo "Building CLI..."
	go build -o $(BUILD_DIR)/$(CLI_BINARY).exe ./cmd/orcacli

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run orchestrator
.PHONY: run-orchestrator
run-orchestrator: build-orchestrator
	@echo "Starting orchestrator..."
	./$(BUILD_DIR)/$(ORCHESTRATOR_BINARY).exe

# Install CLI globally (requires admin privileges)
.PHONY: install
install: build-cli
	@echo "Installing CLI globally..."
	copy $(BUILD_DIR)\$(CLI_BINARY).exe C:\Windows\System32\$(BINARY_NAME).exe

# Uninstall CLI
.PHONY: uninstall
uninstall:
	@echo "Uninstalling CLI..."
	@if exist C:\Windows\System32\$(BINARY_NAME).exe del C:\Windows\System32\$(BINARY_NAME).exe

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run

# Create release builds
.PHONY: release
release: clean
	@echo "Building release binaries..."
	@mkdir $(BUILD_DIR)\windows
	@mkdir $(BUILD_DIR)\linux
	@mkdir $(BUILD_DIR)\darwin
	
	@echo "Building Windows binaries..."
	set GOOS=windows&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/windows/$(ORCHESTRATOR_BINARY).exe ./cmd/orchestrator
	set GOOS=windows&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/windows/$(CLI_BINARY).exe ./cmd/orcacli
	
	@echo "Building Linux binaries..."
	set GOOS=linux&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/linux/$(ORCHESTRATOR_BINARY) ./cmd/orchestrator
	set GOOS=linux&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/linux/$(CLI_BINARY) ./cmd/orcacli
	
	@echo "Building macOS binaries..."
	set GOOS=darwin&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/darwin/$(ORCHESTRATOR_BINARY) ./cmd/orchestrator
	set GOOS=darwin&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/darwin/$(CLI_BINARY) ./cmd/orcacli

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Clean and build all binaries"
	@echo "  clean         - Clean build directory"
	@echo "  build         - Build all binaries"
	@echo "  build-orchestrator - Build orchestrator binary"
	@echo "  build-cli     - Build CLI binary"
	@echo "  deps          - Install dependencies"
	@echo "  test          - Run tests"
	@echo "  run-orchestrator - Build and run orchestrator"
	@echo "  install       - Install CLI globally (requires admin)"
	@echo "  uninstall     - Uninstall CLI"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  release       - Build release binaries for all platforms"
	@echo "  dev-setup     - Setup development environment"
	@echo "  help          - Show this help"