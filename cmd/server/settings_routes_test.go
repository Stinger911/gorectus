package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SettingsHandlersTestSuite is the test suite for settings handlers
type SettingsHandlersTestSuite struct {
	suite.Suite
	mock    sqlmock.Sqlmock
	handler *SettingsHandler
	router  *gin.Engine
}

// SetupSuite runs once before all tests in the suite
func (suite *SettingsHandlersTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

// SetupTest runs before each test
func (suite *SettingsHandlersTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	assert.NoError(suite.T(), err)

	suite.mock = mock

	// Create a mock server
	mockServer := &mockSettingsServerInterface{db: db}
	suite.handler = NewSettingsHandler(mockServer)

	// Setup router
	suite.router = gin.New()
	v1 := suite.router.Group("/api/v1")
	suite.handler.SetupRoutes(v1)
}

// TearDownTest runs after each test
func (suite *SettingsHandlersTestSuite) TearDownTest() {
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// Mock server interface for testing
type mockSettingsServerInterface struct {
	db             *sql.DB
	customAuthFunc gin.HandlerFunc
}

func (m *mockSettingsServerInterface) GetDB() *sql.DB {
	return m.db
}

func (m *mockSettingsServerInterface) AuthMiddleware() gin.HandlerFunc {
	if m.customAuthFunc != nil {
		return m.customAuthFunc
	}
	return func(c *gin.Context) {
		// Default mock authentication - set test user data
		c.Set("user_id", "550e8400-e29b-41d4-a716-446655440000")
		c.Set("user_email", "admin@example.com")
		c.Set("user_role", "Administrator")
		c.Next()
	}
}

func (m *mockSettingsServerInterface) OptionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Status(http.StatusOK)
	}
}

// Helper function to create authenticated request
func (suite *SettingsHandlersTestSuite) createAuthenticatedRequest(method, url string, body interface{}, userID, role string) (*http.Request, *gin.Engine) {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a new router for this request with middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("user_role", role)
		c.Next()
	})

	v1 := router.Group("/api/v1")
	suite.handler.SetupRoutes(v1)

	return req, router
}

// Test GetSettings endpoint
func (suite *SettingsHandlersTestSuite) TestGetSettings_AsAdmin() {
	// Mock the settings query with all new columns
	mockTime := time.Now()
	rows := sqlmock.NewRows([]string{
		"project_name", "project_descriptor", "public_registration", "maintenance_mode",
		"smtp_host", "smtp_port", "smtp_user", "smtp_from_email", "email_enabled",
		"session_timeout", "password_min_length", "require_two_factor", "updated_at",
	}).AddRow("Test Site", "Test Description", true, false, "smtp.example.com", "587", "test@example.com", "noreply@example.com", true, 24, 8, false, mockTime)

	suite.mock.ExpectQuery("SELECT.*FROM settings").WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/settings", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response SettingsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Site", response.Data.SiteName)
	assert.Equal(suite.T(), "Test Description", response.Data.SiteDescription)
	assert.Equal(suite.T(), true, response.Data.AllowRegistration)
}

func (suite *SettingsHandlersTestSuite) TestGetSettings_AsNonAdmin() {
	// Since non-admin access should be denied before database access, we should not expect any database queries

	// Create mock server with custom auth function for non-admin user
	db, _, err := sqlmock.New()
	assert.NoError(suite.T(), err)

	mockServer := &mockSettingsServerInterface{
		db: db,
		customAuthFunc: func(c *gin.Context) {
			c.Set("user_id", "user-id")
			c.Set("user_email", "user@example.com")
			c.Set("user_role", "Public")
			c.Next()
		},
	}
	handler := NewSettingsHandler(mockServer)

	// Setup router
	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.SetupRoutes(v1)

	req := httptest.NewRequest("GET", "/api/v1/settings", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Admin access required", response["error"])
}

// Test UpdateSettings endpoint
func (suite *SettingsHandlersTestSuite) TestUpdateSettings_AsAdmin() {
	// Mock the settings query for getting current settings with all new columns
	mockTime := time.Now()
	rows := sqlmock.NewRows([]string{
		"project_name", "project_descriptor", "public_registration", "maintenance_mode",
		"smtp_host", "smtp_port", "smtp_user", "smtp_from_email", "email_enabled",
		"session_timeout", "password_min_length", "require_two_factor", "updated_at",
	}).AddRow("Old Site", "Old Description", false, false, "", "587", "", "", false, 24, 8, false, mockTime)

	suite.mock.ExpectQuery("SELECT.*FROM settings").WillReturnRows(rows)

	// Mock the update query
	suite.mock.ExpectExec("UPDATE settings SET").WillReturnResult(sqlmock.NewResult(1, 1))

	updateData := UpdateSettingsRequest{
		SiteName:        &[]string{"New Site Name"}[0],
		SiteDescription: &[]string{"New Description"}[0],
	}

	req, router := suite.createAuthenticatedRequest("PATCH", "/api/v1/settings", updateData, "admin-id", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response SettingsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "New Site Name", response.Data.SiteName)
	assert.Equal(suite.T(), "New Description", response.Data.SiteDescription)
}

// Test TestDatabaseConnection endpoint
func (suite *SettingsHandlersTestSuite) TestDatabaseConnection_AsAdmin() {
	// Mock the database ping
	suite.mock.ExpectPing()

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/settings/test-connection", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Database connection successful", response["message"])
}

// Test TestEmailConfiguration endpoint
func (suite *SettingsHandlersTestSuite) TestEmailConfiguration_AsAdmin() {
	// Mock the settings query for email test
	rows := sqlmock.NewRows([]string{"project_name", "project_descriptor", "public_registration", "auth_login_attempts", "updated_at"}).
		AddRow("Test Site", "Test Description", true, 25, "2023-01-01 00:00:00")
	suite.mock.ExpectQuery("SELECT COALESCE\\(project_name").WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/settings/test-email", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Email is not enabled by default, so should return bad request
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Email is not enabled", response["error"])
}

// Run the test suite
func TestSettingsHandlers(t *testing.T) {
	suite.Run(t, new(SettingsHandlersTestSuite))
}
