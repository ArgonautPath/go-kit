.PHONY: help test test-cover test-verbose build clean fmt vet lint examples

# Default target
help:
	@echo "Available targets:"
	@echo "  make test          - Run all tests"
	@echo "  make test-cover    - Run tests with coverage"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make build         - Build all packages"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run go vet"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make examples      - Build example programs"

# Run all tests
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@go test ./... -cover

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	@go test ./... -v

# Build all packages
build:
	@echo "Building packages..."
	@go build ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w . 2>/dev/null || true

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@go clean ./...
	@rm -rf bin/ dist/ coverage.out coverage.html

# Build example programs
examples:
	@echo "Building examples..."
	@mkdir -p bin
	@go build -o bin/logger-demo ./examples/logger-demo
	@go build -o bin/config-demo ./examples/config-demo
	@go build -o bin/httpclient-demo ./examples/httpclient-demo
	@go build -o bin/middleware-demo ./examples/middleware-demo
	@echo "All examples built successfully!"

