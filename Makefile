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
	@echo "  build               - Build the binary"
	@echo "  test                - Run all tests"
	@echo "  test-unit           - Run unit tests only"
	@echo "  test-integration    - Run integration tests only"
	@echo "  clean               - Clean build artifacts"
	@echo "  install             - Install dependencies"
	@echo "  dev                 - Start development environment"
	@echo "  run                 - Run the application locally"
	@echo "  docker-build        - Build Docker image"
	@echo "  docker-run          - Run with Docker Compose"
	@echo "  docker-stop         - Stop Docker services"
	@echo "  lint                - Run linters"
	@echo "  format              - Format code"
	@echo ""
	@echo "Performance Optimization Framework:"
	@echo "  collect-baseline         - Collect performance baseline from Prometheus"
	@echo "  capacity-plan            - Generate capacity planning forecast"
	@echo "  weekly-report            - Generate weekly performance report"
	@echo "  performance-budget-check - Validate against performance budgets"
	@echo "  benchmark-compare        - Compare benchmark results with baseline"
	@echo ""
	@echo "Performance Testing:"
	@echo "  perf-test           - Run all performance tests"
	@echo "  perf-test-frontend  - Run Lighthouse CI tests"
	@echo "  perf-test-api       - Run K6 API load tests"
	@echo "  perf-test-db        - Run database benchmarks"
	@echo "  perf-test-quick     - Quick performance check"
	@echo "  perf-dashboard      - Generate performance dashboard"
	@echo "  perf-clean          - Clean performance artifacts"

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

# Database Migration Management
migrate-status:
	@echo "Checking migration status..."
	@./scripts/migrate.sh status

migrate-up:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make migrate-up VERSION=001"; \
		exit 1; \
	fi
	@./scripts/migrate.sh up $(VERSION)

migrate-down:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make migrate-down VERSION=005"; \
		exit 1; \
	fi
	@./scripts/migrate.sh down $(VERSION)

migrate-rollback-to:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make migrate-rollback-to VERSION=003"; \
		exit 1; \
	fi
	@./scripts/migrate.sh rollback-to $(VERSION)

migrate-dry-run:
	@if [ -z "$(VERSION)" ] || [ -z "$(DIRECTION)" ]; then \
		echo "Usage: make migrate-dry-run DIRECTION=down VERSION=005"; \
		exit 1; \
	fi
	@./scripts/migrate.sh dry-run $(DIRECTION) $(VERSION)

# Generate API documentation (placeholder)
docs:
	@echo "Generating API documentation..."
	# Add documentation generation here

# Performance benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Continuous Performance Optimization targets
collect-baseline:
	@echo "Collecting performance baseline..."
	@mkdir -p baselines
	@cd tools/performance && go run baseline_collector.go \
		--prometheus-url=$${PROMETHEUS_URL:-http://localhost:9090} \
		--lookback=168 \
		--output=../../baselines/baseline_$$(date +%Y%m%d).json

capacity-plan:
	@echo "Generating capacity plan..."
	@mkdir -p reports
	@cd tools/performance && go run capacity_planner.go \
		--input=../../baselines/baseline_latest.json \
		--forecast=90 \
		--scenario=$${SCENARIO:-expected} \
		--output=../../reports/capacity_plan_$$(date +%Y%m%d).json

weekly-report:
	@echo "Generating weekly performance report..."
	@chmod +x scripts/performance/weekly_report.sh
	@./scripts/performance/weekly_report.sh

performance-budget-check:
	@echo "Checking performance budgets..."
	@echo "Not yet implemented - TODO: Add budget validation script"

benchmark-compare:
	@echo "Comparing benchmark results..."
	@echo "Not yet implemented - TODO: Add benchmark comparison"

# Performance testing targets
perf-test: perf-test-frontend perf-test-api perf-test-db
	@echo "All performance tests completed"

perf-test-frontend:
	@echo "Running frontend performance tests with Lighthouse CI..."
	@if command -v lhci > /dev/null; then \
		cd frontend && npm run build && lhci autorun; \
	else \
		echo "Lighthouse CI not installed. Installing..."; \
		npm install -g @lhci/cli && cd frontend && npm run build && lhci autorun; \
	fi

perf-test-api:
	@echo "Running API performance tests with K6..."
	@if command -v k6 > /dev/null; then \
		k6 run tests/performance/api-performance.js; \
	else \
		echo "K6 not installed. Please install from https://k6.io/docs/getting-started/installation/"; \
		exit 1; \
	fi

perf-test-db:
	@echo "Running database performance benchmarks..."
	@cd tests/performance && chmod +x run-benchmarks.sh && ./run-benchmarks.sh

perf-test-quick:
	@echo "Running quick performance check..."
	@k6 run --vus 5 --duration 30s tests/performance/api-performance.js

perf-dashboard:
	@echo "Generating performance dashboard..."
	@python3 tests/performance/performance-dashboard.py

perf-clean:
	@echo "Cleaning performance test artifacts..."
	@rm -rf .lighthouseci
	@rm -rf tests/performance/benchmark-results
	@rm -f performance-results.json
	@rm -f performance-report.html