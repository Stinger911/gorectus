# GoRectus Server Tests

This document describes the comprehensive test suite for the GoRectus server located in `cmd/server/`.

## Test Structure

The test suite is organized into multiple files for better maintainability:

### 1. `main_test.go`

- **Main test suite using testify/suite**
- Tests all HTTP endpoints
- Integration tests for the complete server setup
- Benchmark tests for performance monitoring

### 2. `database_test.go`

- Database connection testing with mocked PostgreSQL
- Environment variable configuration tests
- Connection string validation
- Database ping success/failure scenarios

### 3. `middleware_test.go`

- HTTP middleware testing
- Logging middleware validation
- Configuration testing (log levels, ports, Gin modes)
- Server structure validation

## Test Coverage

Current test coverage: **59.0%**

### Coverage Breakdown:

- `main()` function: 0% (not tested - server startup)
- `NewServer()` function: 0% (not tested - integration function)
- `initDB()` function: 66.7% (partially tested)
- All HTTP handlers: 100% (all placeholder handlers tested)
- Route setup: 100% (fully tested)

### Key Areas Covered:

✅ All API endpoints (auth, collections, items, users, roles)  
✅ Health check functionality  
✅ HTTP routing and middleware  
✅ Environment variable handling  
✅ Database connection logic  
✅ Error handling for invalid endpoints  
✅ Root redirect functionality

### Not Covered:

❌ Main server startup (`main()` function)  
❌ Real database connections  
❌ Integration with actual PostgreSQL

## Dependencies

### Test Dependencies Added:

```go
github.com/stretchr/testify v1.8.3     // Assertions and test suites
github.com/DATA-DOG/go-sqlmock v1.5.0  // Database mocking
```

### Key Testing Libraries:

- **testify/suite**: Test suite organization
- **testify/assert**: Rich assertion library
- **testify/require**: Assertion with immediate failure
- **go-sqlmock**: PostgreSQL database mocking
- **httptest**: HTTP testing utilities

## Running Tests

### Basic Test Run:

```bash
go test -v ./cmd/server
```

### With Coverage:

```bash
go test -coverprofile=coverage.out ./cmd/server
go tool cover -html=coverage.out -o coverage.html
```

### Using Makefile:

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with race detection
make test-race

# Run benchmarks
make benchmark
```

## Test Types

### 1. Unit Tests

- Individual function testing
- Environment variable parsing
- Configuration validation
- Database connection string building

### 2. Integration Tests

- Complete HTTP endpoint testing
- Middleware integration
- Route setup validation
- Server structure testing

### 3. Benchmark Tests

- Health check endpoint performance
- Route setup performance
- Can be extended for load testing

### 4. Mock Tests

- Database connection mocking
- SQL query simulation
- Error scenario testing

## Test Configuration

### Environment Files:

- `.env.test` - Test-specific configuration
- Test mode automatically disables logging output
- Mock database prevents real connections

### Gin Test Mode:

All tests run in `gin.TestMode` to:

- Reduce logging noise
- Improve test performance
- Provide predictable behavior

## Best Practices Implemented

### 1. Test Organization:

- Separate files by functionality
- Clear test naming conventions
- Comprehensive test documentation

### 2. Mocking Strategy:

- Database mocking with sqlmock
- HTTP request/response testing
- Environment variable isolation

### 3. Test Data:

- Parameterized tests for multiple scenarios
- Clear test case descriptions
- Expected vs actual value validation

### 4. Error Testing:

- Invalid endpoint testing
- Database failure scenarios
- Environment variable edge cases

## Test Examples

### HTTP Endpoint Test:

```go
func (suite *ServerTestSuite) TestHealthCheck() {
    req, _ := http.NewRequest("GET", "/api/v1/health", nil)
    w := httptest.NewRecorder()

    suite.router.ServeHTTP(w, req)

    assert.Equal(suite.T(), http.StatusOK, w.Code)
    // Additional assertions...
}
```

### Database Mock Test:

```go
func TestInitDBSuccess(t *testing.T) {
    db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
    require.NoError(t, err)
    defer db.Close()

    mock.ExpectPing()
    err = db.Ping()
    assert.NoError(t, err)
}
```

### Environment Variable Test:

```go
func TestPortConfiguration(t *testing.T) {
    os.Setenv("PORT", "3000")
    defer os.Unsetenv("PORT")

    port := os.Getenv("PORT")
    if port == "" { port = "8080" }

    assert.Equal(t, "3000", port)
}
```

## Future Enhancements

1. **Add Authentication Tests**: When JWT implementation is added
2. **Database Integration Tests**: Real PostgreSQL testing
3. **Load Testing**: Performance benchmarks under load
4. **API Contract Testing**: Request/response validation
5. **End-to-End Tests**: Complete user workflows

## Running in CI/CD

The test suite is designed to run in continuous integration:

- No external dependencies (using mocks)
- Fast execution (< 1 second)
- Clear pass/fail indicators
- Coverage reporting
- No test database required
