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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test suite for user handlers
type UserHandlersTestSuite struct {
	suite.Suite
	handler *UsersHandler
	db      *sql.DB
	mock    sqlmock.Sqlmock
	router  *gin.Engine
}

// SetupSuite runs once before all tests
func (suite *UserHandlersTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	logrus.SetLevel(logrus.PanicLevel)
}

// SetupTest runs before each test
func (suite *UserHandlersTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = db
	suite.mock = mock
	suite.router = gin.New()

	// Create mock server interface
	mockServer := &mockServerInterface{db: db}
	suite.handler = NewUsersHandler(mockServer)

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	suite.handler.SetupRoutes(v1)
}

// TearDownTest runs after each test
func (suite *UserHandlersTestSuite) TearDownTest() {
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
	suite.db.Close()
}

// Mock server interface for testing
type mockServerInterface struct {
	db             *sql.DB
	customAuthFunc gin.HandlerFunc
}

func (m *mockServerInterface) GetDB() *sql.DB {
	return m.db
}

func (m *mockServerInterface) AuthMiddleware() gin.HandlerFunc {
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

func (m *mockServerInterface) OptionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Status(http.StatusOK)
	}
}

// Helper function to create authenticated request
func (suite *UserHandlersTestSuite) createAuthenticatedRequest(method, url string, body interface{}, userID, role string) (*http.Request, *gin.Engine) {
	// Create a new router for this specific test with custom auth middleware
	router := gin.New()

	// Create custom auth middleware for this test
	authMiddleware := func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("user_email", "test@example.com")
		c.Set("user_role", role)
		c.Next()
	}

	// Create mock server interface with custom auth
	mockServer := &mockServerInterface{
		db:             suite.db,
		customAuthFunc: authMiddleware,
	}

	handler := NewUsersHandler(mockServer)

	// Setup routes with the custom auth
	v1 := router.Group("/api/v1")
	handler.SetupRoutes(v1)

	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	return req, router
}

func (suite *UserHandlersTestSuite) TestGetUsers_AsAdmin() {
	// Mock query for getting users
	rows := sqlmock.NewRows([]string{
		"id", "email", "first_name", "last_name", "avatar", "language", "theme",
		"status", "role_id", "role_name", "last_access", "last_page", "provider",
		"external_identifier", "email_notifications", "tags", "created_at", "updated_at",
	}).AddRow(
		"test-id", "test@example.com", "Test", "User", nil, "en-US", "auto",
		"active", "role-id", "User", nil, nil, "default", nil, true, nil,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT u.id, u.email.*FROM users u.*ORDER BY u.created_at DESC").
		WithArgs(50, 0).
		WillReturnRows(rows)

	// Mock count query
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/users", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
	assert.Contains(suite.T(), response, "meta")
}

func (suite *UserHandlersTestSuite) TestGetUsers_AsNonAdmin() {
	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/users", nil, "user-id", "User")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Admin access required", response["error"])
}

func (suite *UserHandlersTestSuite) TestCreateUser_AsAdmin() {
	createReq := CreateUserRequest{
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
		RoleID:    "role-id",
	}

	// Mock check for existing email
	suite.mock.ExpectQuery("SELECT id FROM users WHERE email = \\$1").
		WithArgs("newuser@example.com").
		WillReturnError(sql.ErrNoRows)

	// Mock insert user
	suite.mock.ExpectQuery("INSERT INTO users.*RETURNING id").
		WithArgs(
			"newuser@example.com",
			sqlmock.AnyArg(), // hashed password
			"New",
			"User",
			sqlmock.AnyArg(), // avatar
			"en-US",          // default language
			"auto",           // default theme
			"active",         // default status
			"role-id",
			true, // default email notifications
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-user-id"))

	// Mock fetching created user
	rows := sqlmock.NewRows([]string{
		"id", "email", "first_name", "last_name", "avatar", "language", "theme",
		"status", "role_id", "role_name", "last_access", "last_page", "provider",
		"external_identifier", "email_notifications", "tags", "created_at", "updated_at",
	}).AddRow(
		"new-user-id", "newuser@example.com", "New", "User", nil, "en-US", "auto",
		"active", "role-id", "User", nil, nil, "default", nil, true, nil,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT u.id, u.email.*FROM users u.*WHERE u.id = \\$1").
		WithArgs("new-user-id").
		WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/users", createReq, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
}

func (suite *UserHandlersTestSuite) TestCreateUser_AsNonAdmin() {
	createReq := CreateUserRequest{
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
		RoleID:    "role-id",
	}

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/users", createReq, "user-id", "User")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *UserHandlersTestSuite) TestGetUser_AsAdmin() {
	userID := "target-user-id"

	rows := sqlmock.NewRows([]string{
		"id", "email", "first_name", "last_name", "avatar", "language", "theme",
		"status", "role_id", "role_name", "last_access", "last_page", "provider",
		"external_identifier", "email_notifications", "tags", "created_at", "updated_at",
	}).AddRow(
		userID, "target@example.com", "Target", "User", nil, "en-US", "auto",
		"active", "role-id", "User", nil, nil, "default", nil, true, nil,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT u.id, u.email.*FROM users u.*WHERE u.id = \\$1").
		WithArgs(userID).
		WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/users/"+userID, nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
}

func (suite *UserHandlersTestSuite) TestGetUser_AsSelf() {
	userID := "self-user-id"

	rows := sqlmock.NewRows([]string{
		"id", "email", "first_name", "last_name", "avatar", "language", "theme",
		"status", "role_id", "role_name", "last_access", "last_page", "provider",
		"external_identifier", "email_notifications", "tags", "created_at", "updated_at",
	}).AddRow(
		userID, "self@example.com", "Self", "User", nil, "en-US", "auto",
		"active", "role-id", "User", nil, nil, "default", nil, true, nil,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT u.id, u.email.*FROM users u.*WHERE u.id = \\$1").
		WithArgs(userID).
		WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/users/"+userID, nil, userID, "User")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *UserHandlersTestSuite) TestGetUser_AccessDenied() {
	userID := "target-user-id"

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/users/"+userID, nil, "other-user-id", "User")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *UserHandlersTestSuite) TestUpdateUser_AsSelf() {
	userID := "self-user-id"
	updateReq := UpdateUserRequest{
		FirstName: stringPtr("Updated"),
		LastName:  stringPtr("Name"),
	}

	// Mock user existence check
	suite.mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE id = \\$1\\)").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock update query
	suite.mock.ExpectExec("UPDATE users SET.*WHERE id = \\$3").
		WithArgs("Updated", "Name", userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock fetching updated user
	rows := sqlmock.NewRows([]string{
		"id", "email", "first_name", "last_name", "avatar", "language", "theme",
		"status", "role_id", "role_name", "last_access", "last_page", "provider",
		"external_identifier", "email_notifications", "tags", "created_at", "updated_at",
	}).AddRow(
		userID, "self@example.com", "Updated", "Name", nil, "en-US", "auto",
		"active", "role-id", "User", nil, nil, "default", nil, true, nil,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT u.id, u.email.*FROM users u.*WHERE u.id = \\$1").
		WithArgs(userID).
		WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("PATCH", "/api/v1/users/"+userID, updateReq, userID, "User")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *UserHandlersTestSuite) TestDeleteUser_AsAdmin() {
	userID := "target-user-id"

	// Mock user existence check
	suite.mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE id = \\$1\\)").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock delete query
	suite.mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/users/"+userID, nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User deleted successfully", response["message"])
}

func (suite *UserHandlersTestSuite) TestDeleteUser_AsNonAdmin() {
	userID := "target-user-id"

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/users/"+userID, nil, "user-id", "User")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *UserHandlersTestSuite) TestDeleteUser_SelfDeletion() {
	userID := "admin-id"

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/users/"+userID, nil, userID, "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Cannot delete your own account", response["error"])
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Run the test suite
func TestUserHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlersTestSuite))
}
