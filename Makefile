# Makefile for URL Shortener

.PHONY: build test clean install dev docker-build docker-run docker-stop lint format help

# Variables
BINARY_NAME=url-shortener
MAIN_PATH=./cmd/server/main.go
BUILD_DIR=./bin
DOCKER_IMAGE=url-shortener:latest

# Default target
all: build

# Help target
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  test         - Run all tests"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install dependencies"
	@echo "  dev          - Start development environment"
	@echo "  run          - Run the application locally"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker services"
	@echo "  lint         - Run linters"
	@echo "  format       - Format code"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v ./tests/unit/...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Start development environment (requires Docker)
dev:
	@echo "Starting development environment..."
	@if command -v docker-compose > /dev/null; then \
		docker-compose -f docker/docker-compose.local.yml up -d; \
		echo "Development services started. Run 'make run' to start the application."; \
	else \
		echo "Docker Compose is required for development environment"; \
		exit 1; \
	fi

# Run the application locally
run: build
	@echo "Starting $(BINARY_NAME)..."
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && $(BUILD_DIR)/$(BINARY_NAME); \
	else \
		$(BUILD_DIR)/$(BINARY_NAME); \
	fi

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -f docker/Dockerfile -t $(DOCKER_IMAGE) .

# Run with Docker Compose (production)
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose -f docker/docker-compose.yml up -d

# Stop Docker services
docker-stop:
	@echo "Stopping Docker services..."
	docker-compose -f docker/docker-compose.yml down
	docker-compose -f docker/docker-compose.local.yml down

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	go mod tidy

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Database migrations (placeholder)
migrate-up:
	@echo "Running database migrations..."
	# Add migration command here

migrate-down:
	@echo "Rolling back database migrations..."
	# Add rollback command here

# Generate API documentation (placeholder)
docs:
	@echo "Generating API documentation..."
	# Add documentation generation here

# Performance benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...