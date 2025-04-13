# Variables
BINARY_NAME=fundermapsapp
GO=go

# Default target executed when you just run `make`
.DEFAULT_GOAL := build

# Build the Go application (server)
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build -o $(BINARY_NAME) ./cmd/server

# Run the Go application (depends on build)
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GO) clean
	rm -f $(BINARY_NAME)

# Format Go code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Lint Go code (assumes golangci-lint is installed)
# You might need to install it: https://golangci-lint.run/usage/install/
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Phony targets are targets that don't represent files
.PHONY: build run test clean fmt lint
