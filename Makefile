.PHONY: build run test lint fmt clean help

# Variables
BINARY_NAME=human-release
MAIN_PATH=./cmd
GO_FLAGS=-v

help:
	@echo "Available targets:"
	@echo "  build       - Build the CLI binary"
	@echo "  run         - Build and run the CLI"
	@echo "  test        - Run all tests"
	@echo "  test-single - Run a single test (use TEST=path/to/test)"
	@echo "  lint        - Run linters (golangci-lint)"
	@echo "  fmt         - Format code (go fmt)"
	@echo "  clean       - Remove build artifacts"
	@echo "  install     - Install the CLI binary"

build:
	@echo "Building $(BINARY_NAME)..."
	go build $(GO_FLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

run: build
	./$(BINARY_NAME)

test:
	@echo "Running all tests..."
	go test $(GO_FLAGS) ./...

test-single:
	@if [ -z "$(TEST)" ]; then \
		echo "ERROR: TEST variable not set. Usage: make test-single TEST=./internal/..."; \
		exit 1; \
	fi
	go test $(GO_FLAGS) -run $(TEST) ./...

lint:
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	go clean

install: build
	@echo "Installing $(BINARY_NAME)..."
	go install $(MAIN_PATH)

.DEFAULT_GOAL := help
