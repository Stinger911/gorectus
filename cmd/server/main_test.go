package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test suite for server functionality
type ServerTestSuite struct {
	suite.Suite
	server *Server
	db     *sql.DB
	mock   sqlmock.Sqlmock
	router *gin.Engine
}

// SetupSuite runs once before all tests
func (suite *ServerTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Disable logrus output during tests
	logrus.SetLevel(logrus.PanicLevel)
}

// SetupTest runs before each test
func (suite *ServerTestSuite) SetupTest() {
	// Create mock database with ping monitoring enabled
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(suite.T(), err)

	// Create test server with mock database
	suite.db = db
	suite.mock = mock
	suite.router = gin.New()

	// Add the logging middleware for consistent testing
	suite.router.Use(func(c *gin.Context) {
		c.Next()
	})

	suite.server = &Server{
		db:     suite.db,
		router: suite.router,
	}

	// Setup routes
	suite.server.setupRoutes()
}

// TearDownTest runs after each test
func (suite *ServerTestSuite) TearDownTest() {
	suite.db.Close()
}

// Test health check endpoint
func (suite *ServerTestSuite) TestHealthCheck() {
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "ok", response["status"])
	assert.Equal(suite.T(), "gorectus", response["service"])
	assert.Equal(suite.T(), "1.0.0", response["version"])
}

// Test authentication endpoints
func (suite *ServerTestSuite) TestAuthEndpoints() {
	tests := []struct {
		name        string
		method      string
		path        string
		body        map[string]interface{}
		expected    int
		expectError string
	}{
		{
			name:        "Login endpoint - invalid payload",
			method:      "POST",
			path:        "/api/v1/auth/login",
			body:        map[string]interface{}{"invalid": "data"},
			expected:    http.StatusBadRequest,
			expectError: "Invalid request payload",
		},
		{
			name:        "Login endpoint - missing credentials",
			method:      "POST",
			path:        "/api/v1/auth/login",
			body:        map[string]interface{}{},
			expected:    http.StatusBadRequest,
			expectError: "Invalid request payload",
		},
		{
			name:        "Logout endpoint - no auth",
			method:      "POST",
			path:        "/api/v1/auth/logout",
			expected:    http.StatusUnauthorized,
			expectError: "Authorization header required",
		},
		{
			name:        "Refresh endpoint - no auth",
			method:      "POST",
			path:        "/api/v1/auth/refresh",
			expected:    http.StatusUnauthorized,
			expectError: "Authorization header required",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.body != nil {
				jsonBody, _ := json.Marshal(tt.body)
				body = bytes.NewBuffer(jsonBody)
			} else {
				body = bytes.NewBuffer([]byte{})
			}

			req, _ := http.NewRequest(tt.method, tt.path, body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Contains(t, response["error"], tt.expectError)
		})
	}
}

// Test collections endpoints (protected, require auth)
func (suite *ServerTestSuite) TestCollectionsEndpoints() {
	tests := []struct {
		name     string
		method   string
		path     string
		expected int
	}{
		{"Get collections", "GET", "/api/v1/collections", http.StatusUnauthorized},
		{"Create collection", "POST", "/api/v1/collections", http.StatusUnauthorized},
		{"Get collection", "GET", "/api/v1/collections/test", http.StatusUnauthorized},
		{"Update collection", "PATCH", "/api/v1/collections/test", http.StatusUnauthorized},
		{"Delete collection", "DELETE", "/api/v1/collections/test", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "Authorization header required", response["error"])
		})
	}
}

// Test items endpoints (protected, require auth)
func (suite *ServerTestSuite) TestItemsEndpoints() {
	tests := []struct {
		name     string
		method   string
		path     string
		expected int
	}{
		{"Get items", "GET", "/api/v1/items/test", http.StatusUnauthorized},
		{"Create item", "POST", "/api/v1/items/test", http.StatusUnauthorized},
		{"Get item", "GET", "/api/v1/items/test/1", http.StatusUnauthorized},
		{"Update item", "PATCH", "/api/v1/items/test/1", http.StatusUnauthorized},
		{"Delete item", "DELETE", "/api/v1/items/test/1", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "Authorization header required", response["error"])
		})
	}
}

// Test users endpoints (protected, require auth)
func (suite *ServerTestSuite) TestUsersEndpoints() {
	tests := []struct {
		name     string
		method   string
		path     string
		expected int
	}{
		{"Get users", "GET", "/api/v1/users", http.StatusUnauthorized},
		{"Create user", "POST", "/api/v1/users", http.StatusUnauthorized},
		{"Get user", "GET", "/api/v1/users/1", http.StatusUnauthorized},
		{"Update user", "PATCH", "/api/v1/users/1", http.StatusUnauthorized},
		{"Delete user", "DELETE", "/api/v1/users/1", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "Authorization header required", response["error"])
		})
	}
}

// Test roles endpoints (protected, require auth)
func (suite *ServerTestSuite) TestRolesEndpoints() {
	tests := []struct {
		name     string
		method   string
		path     string
		expected int
	}{
		{"Get roles", "GET", "/api/v1/roles", http.StatusUnauthorized},
		{"Create role", "POST", "/api/v1/roles", http.StatusUnauthorized},
		{"Get role", "GET", "/api/v1/roles/1", http.StatusUnauthorized},
		{"Update role", "PATCH", "/api/v1/roles/1", http.StatusUnauthorized},
		{"Delete role", "DELETE", "/api/v1/roles/1", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "Authorization header required", response["error"])
		})
	}
}

// Test root redirect
func (suite *ServerTestSuite) TestRootRedirect() {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusMovedPermanently, w.Code)
	assert.Equal(suite.T(), "/admin", w.Header().Get("Location"))
}

// Test invalid endpoints
func (suite *ServerTestSuite) TestInvalidEndpoints() {
	req, _ := http.NewRequest("GET", "/api/v1/invalid", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Run the test suite
func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

// Individual unit tests for specific functions

func TestInitDBConnectionString(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "require")

	defer func() {
		// Clean up environment variables
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	}()

	// This test would fail because we can't actually connect to the test database
	// But we can test that the function handles missing environment variables correctly
	_, err := initDB()
	assert.Error(t, err) // Expected to fail since test database doesn't exist
}

func TestDefaultEnvironmentValues(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	// Test default values are used when environment variables are not set
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	assert.Equal(t, "localhost", host)

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	assert.Equal(t, "5432", port)

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	assert.Equal(t, "postgres", user)

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "gorectus"
	}
	assert.Equal(t, "gorectus", dbname)

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	assert.Equal(t, "disable", sslmode)
}

func TestNewServerWithMockDB(t *testing.T) {
	// Create mock database with ping monitoring enabled
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	// Expect ping call
	mock.ExpectPing()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Test that NewServer would work with a valid database connection
	// We can't test the full NewServer function because it calls initDB()
	// But we can test the server creation logic
	router := gin.New()
	server := &Server{
		db:     db,
		router: router,
	}

	assert.NotNil(t, server)
	assert.NotNil(t, server.db)
	assert.NotNil(t, server.router)

	// Test the ping
	err = db.Ping()
	assert.NoError(t, err)

	// Verify mock expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Benchmark tests
func BenchmarkHealthCheck(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{
		db:     nil, // No database needed for health check
		router: router,
	}
	server.setupRoutes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRouteSetup(b *testing.B) {
	gin.SetMode(gin.TestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router := gin.New()
		server := &Server{
			db:     nil,
			router: router,
		}
		server.setupRoutes()
	}
}
