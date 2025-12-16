# Test Utilities Package

This package provides a comprehensive set of testing utilities for the credit management system.

## Features

- **Database Testing**: PostgreSQL test containers with automatic setup and teardown
- **Test Fixtures**: Easy creation of test data for all domain models
- **HTTP Assertions**: Comprehensive assertions for HTTP responses and JSON validation
- **Test Helpers**: Utilities for HTTP requests, file handling, and more

## Installation

To use this package in your service tests:

```bash
go get credit-management/test-utils
```

## Core Components

### 1. Database Testing (database.go)

Set up isolated PostgreSQL test databases using Docker containers:

```go
import testutils "credit-management/test-utils"

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Setup test database
    testDB, err := testutils.SetupTestDatabase(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer testDB.Teardown(ctx)

    // Run tests
    code := m.Run()
    os.Exit(code)
}

func TestSomething(t *testing.T) {
    // Clean database before test
    testDB.CleanDatabase()

    // Use database
    db := testDB.GetConnection()
    // ... your test code
}
```

### 2. Test Fixtures (fixtures.go)

Create test data easily with the FixtureBuilder:

```go
fb := testutils.NewFixtureBuilder(db)

// Create a user
user, err := fb.CreateUser(map[string]interface{}{
    "username": "testuser",
    "email": "test@example.com",
    "role": "student",
})

// Create an activity
activity, err := fb.CreateActivity(map[string]interface{}{
    "name": "Test Workshop",
    "organizer_id": user.ID,
    "status": "published",
})

// Create a participant
participant, err := fb.CreateParticipant(activity.ID, user.ID, nil)

// Create multiple users at once
users, err := fb.CreateMultipleUsers(10, map[string]interface{}{
    "department": "Computer Science",
})

// Setup a complete test scenario
scenario, err := fb.SetupBasicScenario()
organizer := scenario["organizer"].(*testutils.User)
student1 := scenario["student1"].(*testutils.User)
activity1 := scenario["activity1"].(*testutils.Activity)
```

### 3. Assertions (assertions.go)

Comprehensive HTTP and JSON response assertions:

```go
ah := testutils.NewAssertHelper(t)

// HTTP status assertions
ah.AssertHTTPStatus(response, http.StatusOK)
ah.AssertSuccessResponse(response)

// JSON assertions
ah.AssertJSONFieldEquals(response, "data.username", "testuser")
ah.AssertJSONFieldExists(response, "data.id")
ah.AssertJSONArrayLength(response, "data.activities", 5)

// Error assertions
ah.AssertErrorResponse(response, http.StatusBadRequest, "invalid email")
ah.AssertValidationError(response, "email")

// Pagination assertions
ah.AssertPaginationResponse(response, 100)

// Other assertions
ah.AssertEqual(expected, actual)
ah.AssertNotNil(user)
ah.RequireNoError(err) // Stops test on failure
```

### 4. HTTP Request Helpers (helpers.go)

Create HTTP requests for testing:

```go
// JSON request
req, err := testutils.CreateJSONRequest("POST", "/api/users", map[string]interface{}{
    "username": "testuser",
    "email": "test@example.com",
})

// Add authentication
testutils.AddAuthHeader(req, "your-jwt-token")

// Perform request
router := setupRouter()
resp := testutils.PerformRequest(router, req)

// Parse response
var result map[string]interface{}
testutils.ParseJSONResponse(resp, &result)

// Multipart form request
req, err := testutils.CreateMultipartRequestWithBytes(
    "POST",
    "/api/upload",
    map[string]string{"field": "value"},
    map[string][]byte{"file": fileBytes},
)
```

### 5. Utility Functions (helpers.go)

Various helpful utilities:

```go
// Generate IDs
id := testutils.GenerateID()
shortID := testutils.GenerateShortID()

// Random data
email := testutils.RandomEmail()
username := testutils.RandomUsername()
str := testutils.RandomString(10)

// Temporary files
filePath, err := testutils.CreateTempFile(content, "test.pdf")
defer testutils.CleanupTempFile(filePath)

// Time mocking
mockTime := testutils.NewMockTime(time.Now())
mockTime.Advance(24 * time.Hour)
currentTime := mockTime.Now()

// Retry logic
err := testutils.Retry(func() error {
    return performOperation()
}, 3, time.Second)

// Wait for condition
success := testutils.WaitForCondition(func() bool {
    return isReady()
}, 10*time.Second, 100*time.Millisecond)

// Environment setup
cleanup := testutils.SetupTestEnv(map[string]string{
    "DATABASE_URL": "test-db-url",
    "JWT_SECRET": "test-secret",
})
defer cleanup()
```

## Complete Test Example

```go
package tests

import (
    "context"
    "net/http"
    "testing"

    testutils "credit-management/test-utils"
)

var testDB *testutils.TestDatabase

func TestMain(m *testing.M) {
    ctx := context.Background()
    var err error

    testDB, err = testutils.SetupTestDatabase(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer testDB.Teardown(ctx)

    code := m.Run()
    os.Exit(code)
}

func TestUserRegistration(t *testing.T) {
    // Setup
    testDB.CleanDatabase()
    ah := testutils.NewAssertHelper(t)
    router := setupRouter(testDB.DB)

    // Create request
    req, err := testutils.CreateJSONRequest("POST", "/api/register", map[string]interface{}{
        "username": "newuser",
        "email": "newuser@example.com",
        "password": "securepass123",
    })
    ah.RequireNoError(err)

    // Perform request
    resp := testutils.PerformRequest(router, req)

    // Assert response
    ah.AssertHTTPStatus(resp, http.StatusCreated)
    ah.AssertJSONFieldExists(resp, "data.id")
    ah.AssertJSONFieldEquals(resp, "data.username", "newuser")
    ah.AssertJSONFieldEquals(resp, "data.email", "newuser@example.com")
}

func TestActivityCreation(t *testing.T) {
    // Setup
    testDB.CleanDatabase()
    fb := testutils.NewFixtureBuilder(testDB.DB)
    ah := testutils.NewAssertHelper(t)

    // Create test data
    organizer, err := fb.CreateUser(map[string]interface{}{
        "role": "teacher",
    })
    ah.RequireNoError(err)

    router := setupRouter(testDB.DB)

    // Create activity request
    req, err := testutils.CreateJSONRequest("POST", "/api/activities", map[string]interface{}{
        "name": "Test Workshop",
        "type": "workshop",
        "organizer_id": organizer.ID,
    })
    ah.RequireNoError(err)

    // Add authentication
    token := generateTestToken(organizer.ID)
    testutils.AddAuthHeader(req, token)

    // Perform request
    resp := testutils.PerformRequest(router, req)

    // Assert response
    ah.AssertSuccessResponse(resp)
    ah.AssertJSONFieldEquals(resp, "data.name", "Test Workshop")
}
```

## Best Practices·

1. **Clean Database Between Tests**: Always call `testDB.CleanDatabase()` to ensure test isolation
2. **Use Fixtures for Complex Setup**: Use FixtureBuilder for creating related test data
3. **Use RequireNoError for Critical Checks**: Use `RequireNoError` to stop tests early if setup fails
4. **Create Scenario Methods**: Add custom scenario setup methods to FixtureBuilder for common test cases
5. **Test with Docker Running**: Ensure Docker is running for testcontainers to work

## Environment Requirements

- Go 1.24+
- Docker (for test containers)
- PostgreSQL compatible database

## Dependencies

- github.com/gin-gonic/gin - HTTP framework
- github.com/stretchr/testify - Assertions
- github.com/testcontainers/testcontainers-go - Test containers
- gorm.io/gorm - Database ORM
- gorm.io/driver/postgres - PostgreSQL driver

## Troubleshooting

### Docker Connection Issues

If you get Docker connection errors, ensure:
- Docker daemon is running
- Your user has Docker permissions
- Docker socket is accessible

### Database Connection Timeouts

If database connections timeout:
- Increase the timeout in `SetupTestDatabase`
- Check Docker resource allocation
- Ensure no port conflicts on 5432

### Test Isolation Issues

If tests interfere with each other:
- Call `CleanDatabase()` in test setup
- Use unique IDs for test data
- Run tests sequentially if needed

## Contributing

When adding new utilities:
1. Add comprehensive documentation
2. Include usage examples
3. Write unit tests for the utility itself
4. Update this README
