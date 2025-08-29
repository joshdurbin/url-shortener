# Makefile for URL Shortener

.PHONY: build test test-unit test-integration clean run-server help install-tools generate fmt lint

# Binary name
BINARY_NAME=url-shortener

# Build the application
build:
	go build -o $(BINARY_NAME) ./cmd/server

# Install the binary to GOPATH/bin
install:
	go install ./cmd/server

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	go test -v -race -short ./internal/...

# Run integration tests
test-integration:
	go test -v -race -run Integration ./tests/integration/...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -f *.db

# Run the server
run-server: build
	./$(BINARY_NAME) server

# Install required tools
install-tools:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate code (sqlc, mocks, etc.)
generate:
	$(shell go env GOPATH)/bin/sqlc generate

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Development setup
dev-setup: install-tools tidy generate fmt

# Run benchmarks
bench:
	go test -bench=. -benchmem ./internal/...

# Run server with custom settings
run-example: build
	./$(BINARY_NAME) server --port 9000 --db-path example.db --metrics-port 9091

# Docker build (if Dockerfile exists)
docker-build:
	docker build -t url-shortener .

# Show help
help:
	@echo "Available commands:"
	@echo "  build              Build the application"
	@echo "  install            Install the binary"
	@echo "  test               Run all tests"
	@echo "  test-unit          Run unit tests only"
	@echo "  test-integration   Run integration tests only"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  clean              Clean build artifacts"
	@echo "  run-server         Build and run the server"
	@echo "  install-tools      Install development tools"
	@echo "  generate           Generate code (sqlc, etc.)"
	@echo "  fmt                Format code"
	@echo "  lint               Lint code"
	@echo "  tidy               Tidy dependencies"
	@echo "  dev-setup          Set up development environment"
	@echo "  bench              Run benchmarks"
	@echo "  run-example        Run server with custom settings"
	@echo "  help               Show this help message"