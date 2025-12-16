# Credit Management System - Test Suite

.PHONY: help test test-unit test-integration test-coverage test-all clean install-deps

# Default target
help:
	@echo "Available targets:"
	@echo "  make install-deps      - Install test dependencies"
	@echo "  make test             - Run all tests"
	@echo "  make test-unit        - Run unit tests only"
	@echo "  make test-integration - Run integration tests only"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make test-auth        - Run auth-service tests"
	@echo "  make test-activity    - Run credit-activity-service tests"
	@echo "  make test-user        - Run user-service tests"
	@echo "  make clean            - Clean test artifacts"

# Install test dependencies
install-deps:
	@echo "Installing test dependencies..."
	cd test-utils && go mod tidy && go mod download
	cd auth-service && go mod tidy && go mod download
	cd credit-activity-service && go mod tidy && go mod download
	cd user-service && go mod tidy && go mod download
	@echo "Dependencies installed successfully"

# Run all tests
test:
	@echo "Running all tests..."
	@$(MAKE) test-auth
	@$(MAKE) test-activity
	@echo "All tests completed!"

# Run unit tests only (fast, no database)
test-unit:
	@echo "Running unit tests..."
	cd auth-service && go test -v -short ./...
	cd credit-activity-service && go test -v -short ./...
	cd user-service && go test -v -short ./...

# Run integration tests only (with database)
test-integration:
	@echo "Running integration tests..."
	@echo "Checking Docker..."
	@docker info > /dev/null 2>&1 || (echo "Error: Docker is not running" && exit 1)
	cd auth-service/tests && go test -v -timeout 10m ./...
	cd credit-activity-service/tests && go test -v -timeout 10m ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	cd auth-service && go test -v -coverprofile=../coverage/auth-coverage.out ./...
	cd credit-activity-service && go test -v -coverprofile=../coverage/activity-coverage.out ./...
	cd user-service && go test -v -coverprofile=../coverage/user-coverage.out ./...
	@echo "\nCoverage Summary:"
	@go tool cover -func=coverage/auth-coverage.out | grep total || true
	@go tool cover -func=coverage/activity-coverage.out | grep total || true
	@go tool cover -func=coverage/user-coverage.out | grep total || true

# Generate HTML coverage reports
test-coverage-html: test-coverage
	@echo "Generating HTML coverage reports..."
	cd auth-service && go tool cover -html=../coverage/auth-coverage.out -o ../coverage/auth-coverage.html
	cd credit-activity-service && go tool cover -html=../coverage/activity-coverage.out -o ../coverage/activity-coverage.html
	cd user-service && go tool cover -html=../coverage/user-coverage.out -o ../coverage/user-coverage.html
	@echo "Coverage reports generated in coverage/ directory"

# Run auth-service tests
test-auth:
	@echo "Running auth-service tests..."
	@docker info > /dev/null 2>&1 || (echo "Error: Docker is not running" && exit 1)
	cd auth-service/tests && go test -v -timeout 10m ./...

# Run credit-activity-service tests
test-activity:
	@echo "Running credit-activity-service tests..."
	@docker info > /dev/null 2>&1 || (echo "Error: Docker is not running" && exit 1)
	cd credit-activity-service/tests && go test -v -timeout 10m ./...

# Run user-service tests
test-user:
	@echo "Running user-service tests..."
	@docker info > /dev/null 2>&1 || (echo "Error: Docker is not running" && exit 1)
	cd user-service/tests && go test -v -timeout 10m ./...

# Run tests in parallel (faster but uses more resources)
test-parallel:
	@echo "Running tests in parallel..."
	@docker info > /dev/null 2>&1 || (echo "Error: Docker is not running" && exit 1)
	cd auth-service/tests && go test -v -timeout 10m -parallel 4 ./... &
	cd credit-activity-service/tests && go test -v -timeout 10m -parallel 4 ./... &
	wait

# Run specific test by name
test-specific:
	@if [ -z "$(SERVICE)" ] || [ -z "$(TEST)" ]; then \
		echo "Usage: make test-specific SERVICE=auth-service TEST=TestLoginWithUsername"; \
		exit 1; \
	fi
	@echo "Running $(TEST) in $(SERVICE)..."
	cd $(SERVICE)/tests && go test -v -run $(TEST) ./...

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	rm -rf coverage/
	find . -name "*.test" -delete
	find . -name "*.out" -delete
	@echo "Clean complete"

# Verify test setup
verify-setup:
	@echo "Verifying test setup..."
	@echo "Checking Docker..."
	@docker info > /dev/null 2>&1 && echo "✓ Docker is running" || echo "✗ Docker is not running"
	@echo "Checking Go installation..."
	@go version && echo "✓ Go is installed" || echo "✗ Go is not installed"
	@echo "Checking test-utils..."
	@[ -d "test-utils" ] && echo "✓ test-utils exists" || echo "✗ test-utils not found"
	@echo "Checking test directories..."
	@[ -d "auth-service/tests" ] && echo "✓ auth-service/tests exists" || echo "✗ auth-service/tests not found"
	@[ -d "credit-activity-service/tests" ] && echo "✓ credit-activity-service/tests exists" || echo "✗ credit-activity-service/tests not found"

# Run tests with verbose output and no cache
test-verbose:
	@echo "Running tests with verbose output..."
	cd auth-service/tests && go test -v -count=1 -timeout 10m ./...
	cd credit-activity-service/tests && go test -v -count=1 -timeout 10m ./...

# Run tests and watch for changes (requires entr or similar tool)
test-watch:
	@echo "Watching for file changes..."
	@which entr > /dev/null || (echo "Error: 'entr' is not installed. Install with: apt-get install entr" && exit 1)
	find . -name "*.go" | entr -c make test

# Benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	cd auth-service && go test -bench=. -benchmem ./...
	cd credit-activity-service && go test -bench=. -benchmem ./...
