## Testing Guide

This document provides comprehensive information about the testing infrastructure for the Credit Management System.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Test Structure](#test-structure)
4. [Running Tests](#running-tests)
5. [Writing Tests](#writing-tests)
6. [Coverage Requirements](#coverage-requirements)
7. [CI/CD Integration](#cicd-integration)
8. [Troubleshooting](#troubleshooting)

## Overview

The Credit Management System has a comprehensive test suite that includes:

- **Unit Tests**: Fast tests for individual functions and components
- **Integration Tests**: Tests for API endpoints and database interactions
- **Security Tests**: Tests for authentication, authorization, and file upload security
- **End-to-End Tests**: Complete workflow tests across the entire lifecycle

### Test Coverage Goals

- **Minimum Coverage**: 60% code coverage across all services
- **Critical Paths**: 80%+ coverage for authentication and credit calculation
- **Security Features**: 100% coverage for security-critical code

## Quick Start

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for test containers)
- Make (optional, but recommended)

### Running All Tests

```bash
# Using make
make test

# Using test script
./run-tests.sh all

# Using go directly
cd auth-service/tests && go test -v ./...
cd credit-activity-service/tests && go test -v ./...
```

### Running Tests with Coverage

```bash
# Using make
make test-coverage

# View HTML coverage report
make test-coverage-html
open coverage/auth-coverage.html
```

## Test Structure

```
credit-management/
├── test-utils/                 # Shared testing utilities
│   ├── database.go            # Test database setup with containers
│   ├── fixtures.go            # Test data builders
│   ├── assertions.go          # Custom assertions
│   ├── helpers.go             # HTTP and utility helpers
│   └── README.md              # Detailed test-utils documentation
├── auth-service/
│   └── tests/
│       └── auth_test.go       # Authentication tests
├── credit-activity-service/
│   └── tests/
│       ├── activity_test.go   # Activity lifecycle tests
│       ├── participant_test.go # Participant management tests
│       └── attachment_test.go # File upload security tests
├── .github/
│   └── workflows/
│       └── test.yml           # CI/CD test automation
├── Makefile                   # Test automation tasks
└── run-tests.sh              # Test runner script
```

## Running Tests

### Using Make (Recommended)

```bash
# Run all tests
make test

# Run specific service tests
make test-auth
make test-activity
make test-user

# Run with coverage
make test-coverage
make test-coverage-html

# Run specific test
make test-specific SERVICE=auth-service TEST=TestLoginWithUsername

# Clean test artifacts
make clean
```

### Using Test Script

```bash
# Run all tests
./run-tests.sh all

# Run specific service
./run-tests.sh auth
./run-tests.sh activity

# Run with coverage
./run-tests.sh coverage

# Run specific test
./run-tests.sh specific -s auth-service -t TestLoginWithUsername

# Verify test setup
./run-tests.sh verify
```

### Using Go Commands Directly

```bash
# Run tests for a specific service
cd auth-service/tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestLoginWithUsername ./...

# Run tests with race detector
go test -v -race ./...

# Run benchmarks
go test -v -bench=. ./...
```

## Writing Tests

### Test File Naming

- Test files must end with `_test.go`
- Place tests in a `tests/` subdirectory
- Name test functions with `Test` prefix: `func TestFeatureName(t *testing.T)`

### Example Test Structure

```go
package tests

import (
    "testing"
    "net/http"

    testutils "credit-management/test-utils"
)

func TestExample(t *testing.T) {
    // 1. Setup
    testDB.CleanDatabase("table1", "table2")
    ah := testutils.NewAssertHelper(t)

    // 2. Create test data
    user, err := createTestUser(t, nil)
    ah.RequireNoError(err)

    // 3. Make request
    req, err := testutils.CreateJSONRequest("POST", "/api/endpoint", payload)
    ah.RequireNoError(err)

    resp := testutils.PerformRequest(testRouter, req)

    // 4. Assert results
    ah.AssertHTTPStatus(resp, http.StatusOK)
    ah.AssertJSONFieldEquals(resp, "data.field", "expected")

    // 5. Verify database state
    var record Model
    testDB.DB.First(&record, "id = ?", id)
    ah.AssertEqual("expected", record.Field)
}
```

### Using Test Utilities

#### Database Setup

```go
// In TestMain
func TestMain(m *testing.M) {
    ctx := context.Background()
    var err error

    testDB, err = testutils.SetupTestDatabase(ctx)
    if err != nil {
        panic(err)
    }
    defer testDB.Teardown(ctx)

    code := m.Run()
    os.Exit(code)
}

// Clean database before each test
func TestSomething(t *testing.T) {
    testDB.CleanDatabase("users", "activities")
    // ... test code
}
```

#### Creating Test Data

```go
// Using FixtureBuilder
fb := testutils.NewFixtureBuilder(testDB.DB)

user, err := fb.CreateUser(map[string]interface{}{
    "username": "testuser",
    "email": "test@example.com",
})

activity, err := fb.CreateActivity(map[string]interface{}{
    "name": "Test Workshop",
    "organizer_id": user.ID,
})

// Create complete scenario
scenario, err := fb.SetupBasicScenario()
organizer := scenario["organizer"].(*testutils.User)
```

#### HTTP Assertions

```go
ah := testutils.NewAssertHelper(t)

// Status assertions
ah.AssertHTTPStatus(resp, http.StatusOK)
ah.AssertSuccessResponse(resp)

// JSON assertions
ah.AssertJSONFieldEquals(resp, "data.username", "testuser")
ah.AssertJSONFieldExists(resp, "data.id")
ah.AssertJSONArrayLength(resp, "data.items", 5)

// Error assertions
ah.AssertErrorResponse(resp, http.StatusBadRequest, "invalid email")
ah.AssertValidationError(resp, "email")
```

### Test Categories

#### Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name        string
        input       interface{}
        expectError bool
        errorMsg    string
    }{
        {
            name:        "valid input",
            input:       validData,
            expectError: false,
        },
        {
            name:        "invalid email",
            input:       invalidEmailData,
            expectError: true,
            errorMsg:    "invalid email",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

#### Concurrent Tests

```go
func TestConcurrentOperations(t *testing.T) {
    const numGoroutines = 10
    results := make(chan error, numGoroutines)

    for i := 0; i < numGoroutines; i++ {
        go func() {
            // Concurrent operation
            results <- performOperation()
        }()
    }

    for i := 0; i < numGoroutines; i++ {
        err := <-results
        assert.NoError(t, err)
    }
}
```

## Coverage Requirements

### Measuring Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage by function
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Thresholds

- **Overall**: Minimum 60% coverage
- **Critical Paths**: 80%+ coverage
  - Authentication (login, token validation)
  - Authorization (permission checks)
  - Credit calculation
  - File upload security
- **New Code**: All new code must have tests

### Enforcing Coverage in CI

The CI pipeline automatically checks coverage and fails if below threshold:

```yaml
- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$COVERAGE < 60" | bc -l) )); then
      echo "Coverage below threshold"
      exit 1
    fi
```

## CI/CD Integration

### GitHub Actions

Tests run automatically on:
- Push to `master`, `main`, or `develop` branches
- Pull requests to these branches
- Manual workflow dispatch

### Workflow Jobs

1. **test-auth-service**: Runs auth service tests with PostgreSQL and Redis
2. **test-activity-service**: Runs activity service tests with PostgreSQL
3. **test-integration**: Runs cross-service integration tests
4. **lint**: Runs golangci-lint for code quality
5. **security-scan**: Runs Gosec security scanner
6. **test-summary**: Aggregates results

### Viewing Results

- Check the "Actions" tab in GitHub
- Coverage reports are uploaded to Codecov
- Failed tests show detailed error messages

## Troubleshooting

### Docker Not Running

**Error**: "Cannot connect to Docker daemon"

**Solution**:
```bash
# Start Docker
sudo systemctl start docker

# Verify Docker is running
docker info
```

### Port Conflicts

**Error**: "Address already in use"

**Solution**:
```bash
# Find process using port
lsof -i :5432

# Kill process or use different port in test setup
```

### Test Database Connection Issues

**Error**: "Failed to connect to database"

**Solution**:
```bash
# Check Docker container logs
docker ps
docker logs <container-id>

# Increase connection timeout in test setup
# Wait for database to be ready before running tests
```

### Slow Tests

Tests taking too long? Try:

```bash
# Run tests in parallel
go test -parallel 4 ./...

# Run specific tests
go test -run TestSpecificTest ./...

# Use test caching
go test ./...  # Second run uses cache
```

### Clean Test State

If tests are interfering with each other:

```bash
# Clean all test artifacts
make clean

# Ensure database is cleaned before each test
testDB.CleanDatabase()

# Use isolated test containers
```

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Always clean database state before tests
3. **Fast Tests**: Keep tests fast by avoiding unnecessary sleeps
4. **Clear Names**: Use descriptive test function names
5. **Table-Driven**: Use table-driven tests for multiple scenarios
6. **Error Messages**: Provide clear assertion messages
7. **Coverage**: Aim for high coverage on critical paths
8. **Documentation**: Document complex test scenarios

## Additional Resources

- [Test Utils README](test-utils/README.md) - Detailed utility documentation
- [Go Testing Package](https://pkg.go.dev/testing) - Official Go testing docs
- [Testcontainers](https://golang.testcontainers.org/) - Container setup docs
- [testify](https://github.com/stretchr/testify) - Assertion library docs

## Getting Help

If you encounter issues:

1. Check this documentation
2. Review existing tests for examples
3. Check GitHub Actions logs for CI failures
4. Ask the team in Slack #testing channel
