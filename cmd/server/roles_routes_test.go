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
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test suite for role handlers
type RoleHandlersTestSuite struct {
	suite.Suite
	handler *RolesHandler
	db      *sql.DB
	mock    sqlmock.Sqlmock
	router  *gin.Engine
}

// SetupSuite runs once before all tests
func (suite *RoleHandlersTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	logrus.SetLevel(logrus.PanicLevel)
}

// SetupTest runs before each test
func (suite *RoleHandlersTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = db
	suite.mock = mock
	suite.router = gin.New()

	// Create mock server interface
	mockServer := &mockRoleServerInterface{db: db}
	suite.handler = NewRolesHandler(mockServer)

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	suite.handler.SetupRoutes(v1)
}

// TearDownTest runs after each test
func (suite *RoleHandlersTestSuite) TearDownTest() {
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
	suite.db.Close()
}

// Mock server interface for testing
type mockRoleServerInterface struct {
	db             *sql.DB
	customAuthFunc gin.HandlerFunc
}

func (m *mockRoleServerInterface) GetDB() *sql.DB {
	return m.db
}

func (m *mockRoleServerInterface) AuthMiddleware() gin.HandlerFunc {
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

func (m *mockRoleServerInterface) OptionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Status(http.StatusOK)
	}
}

// Helper function to create authenticated request
func (suite *RoleHandlersTestSuite) createAuthenticatedRequest(method, url string, body interface{}, userID, role string) (*http.Request, *gin.Engine) {
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
	mockServer := &mockRoleServerInterface{
		db:             suite.db,
		customAuthFunc: authMiddleware,
	}

	handler := NewRolesHandler(mockServer)

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

// Test GetRoles endpoint
func (suite *RoleHandlersTestSuite) TestGetRoles_AsAdmin() {
	// Mock query for getting roles - using pq.Array for ip_access
	rows := sqlmock.NewRows([]string{
		"id", "name", "icon", "description", "ip_access", "enforce_tfa", "admin_access", "app_access",
		"created_at", "updated_at",
	}).AddRow(
		"role-id-1", "Administrator", "verified_user", "System administrator", pq.Array([]string{}), false, true, true,
		time.Now(), time.Now(),
	).AddRow(
		"role-id-2", "Editor", "edit", "Content editor", pq.Array([]string{}), false, false, true,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,.*FROM roles.*ORDER BY created_at DESC").
		WithArgs(50, 0).
		WillReturnRows(rows)

	// Mock count query
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM roles").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/roles", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
	assert.Contains(suite.T(), response, "meta")

	// Check that we have roles in the response
	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)

	// Check meta information
	meta := response["meta"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), meta["page"])
	assert.Equal(suite.T(), float64(50), meta["limit"])
	assert.Equal(suite.T(), float64(2), meta["total"])
}

func (suite *RoleHandlersTestSuite) TestGetRoles_AsNonAdmin() {
	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/roles", nil, "user-id", "Editor")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Admin access required", response["error"])
}

// Test CreateRole endpoint
func (suite *RoleHandlersTestSuite) TestCreateRole_AsAdmin() {
	createReq := CreateRoleRequest{
		Name:        "Editor",
		Icon:        "edit",
		Description: stringPtr("Content editor role"),
		IPAccess:    []string{},
		EnforceTFA:  boolPtr(false),
		AdminAccess: boolPtr(false),
		AppAccess:   boolPtr(true),
	}

	// Mock check for existing role name
	suite.mock.ExpectQuery("SELECT id FROM roles WHERE name = \\$1").
		WithArgs("Editor").
		WillReturnError(sql.ErrNoRows)

	// Mock insert role
	suite.mock.ExpectQuery("INSERT INTO roles.*RETURNING id").
		WithArgs(
			"Editor",
			"edit",
			stringPtr("Content editor role"),
			pq.Array([]string{}),
			false, // enforce_tfa
			false, // admin_access
			true,  // app_access
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-role-id"))

	// Mock fetching created role
	rows := sqlmock.NewRows([]string{
		"id", "name", "icon", "description", "ip_access", "enforce_tfa", "admin_access", "app_access",
		"created_at", "updated_at",
	}).AddRow(
		"new-role-id", "Editor", "edit", "Content editor role", pq.Array([]string{}), false, false, true,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,.*FROM roles.*WHERE id = \\$1").
		WithArgs("new-role-id").
		WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/roles", createReq, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Editor", data["name"])
	assert.Equal(suite.T(), "edit", data["icon"])
}

func (suite *RoleHandlersTestSuite) TestCreateRole_AsNonAdmin() {
	createReq := CreateRoleRequest{
		Name: "Editor",
	}

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/roles", createReq, "user-id", "Editor")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *RoleHandlersTestSuite) TestCreateRole_DuplicateName() {
	createReq := CreateRoleRequest{
		Name: "Administrator",
	}

	// Mock check for existing role name - returns a row (role exists)
	suite.mock.ExpectQuery("SELECT id FROM roles WHERE name = \\$1").
		WithArgs("Administrator").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("existing-role-id"))

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/roles", createReq, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Role name already exists", response["error"])
}

// Test GetRole endpoint
func (suite *RoleHandlersTestSuite) TestGetRole_AsAdmin() {
	roleID := "test-role-id"

	rows := sqlmock.NewRows([]string{
		"id", "name", "icon", "description", "ip_access", "enforce_tfa", "admin_access", "app_access",
		"created_at", "updated_at",
	}).AddRow(
		roleID, "Editor", "edit", "Content editor", pq.Array([]string{"192.168.1.0/24"}), false, false, true,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,.*FROM roles.*WHERE id = \\$1").
		WithArgs(roleID).
		WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/roles/"+roleID, nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Editor", data["name"])
	assert.Equal(suite.T(), "edit", data["icon"])
}

func (suite *RoleHandlersTestSuite) TestGetRole_AsNonAdmin() {
	roleID := "test-role-id"

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/roles/"+roleID, nil, "user-id", "Editor")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *RoleHandlersTestSuite) TestGetRole_NotFound() {
	roleID := "non-existent-role-id"

	suite.mock.ExpectQuery("SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,.*FROM roles.*WHERE id = \\$1").
		WithArgs(roleID).
		WillReturnError(sql.ErrNoRows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/roles/"+roleID, nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Role not found", response["error"])
}

// Test DeleteRole endpoint
func (suite *RoleHandlersTestSuite) TestDeleteRole_AsAdmin() {
	roleID := "test-role-id"

	// Mock fetching existing role
	existingRows := sqlmock.NewRows([]string{
		"id", "name", "icon", "description", "ip_access", "enforce_tfa", "admin_access", "app_access",
		"created_at", "updated_at",
	}).AddRow(
		roleID, "Editor", "edit", "Content editor", pq.Array([]string{}), false, false, true,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,.*FROM roles.*WHERE id = \\$1").
		WithArgs(roleID).
		WillReturnRows(existingRows)

	// Mock check for users assigned to this role
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE role_id = \\$1").
		WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock delete query
	suite.mock.ExpectExec("DELETE FROM roles WHERE id = \\$1").
		WithArgs(roleID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/roles/"+roleID, nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Role deleted successfully", response["message"])
}

func (suite *RoleHandlersTestSuite) TestDeleteRole_AsNonAdmin() {
	roleID := "test-role-id"

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/roles/"+roleID, nil, "user-id", "Editor")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *RoleHandlersTestSuite) TestDeleteRole_SystemRoleProtection() {
	roleID := "admin-role-id"

	// Mock fetching existing Administrator role
	existingRows := sqlmock.NewRows([]string{
		"id", "name", "icon", "description", "ip_access", "enforce_tfa", "admin_access", "app_access",
		"created_at", "updated_at",
	}).AddRow(
		roleID, "Administrator", "verified_user", "System administrator", pq.Array([]string{}), false, true, true,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,.*FROM roles.*WHERE id = \\$1").
		WithArgs(roleID).
		WillReturnRows(existingRows)

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/roles/"+roleID, nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Cannot delete system roles", response["error"])
}

func (suite *RoleHandlersTestSuite) TestDeleteRole_RoleInUse() {
	roleID := "test-role-id"

	// Mock fetching existing role
	existingRows := sqlmock.NewRows([]string{
		"id", "name", "icon", "description", "ip_access", "enforce_tfa", "admin_access", "app_access",
		"created_at", "updated_at",
	}).AddRow(
		roleID, "Editor", "edit", "Content editor", pq.Array([]string{}), false, false, true,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,.*FROM roles.*WHERE id = \\$1").
		WithArgs(roleID).
		WillReturnRows(existingRows)

	// Mock check for users assigned to this role - return count > 0
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE role_id = \\$1").
		WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/roles/"+roleID, nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Cannot delete role that is assigned to users", response["error"])
	assert.Equal(suite.T(), float64(5), response["users_count"])
}

func (suite *RoleHandlersTestSuite) TestDeleteRole_RoleNotFound() {
	roleID := "non-existent-role-id"

	suite.mock.ExpectQuery("SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,.*FROM roles.*WHERE id = \\$1").
		WithArgs(roleID).
		WillReturnError(sql.ErrNoRows)

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/roles/"+roleID, nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

// Run the test suite
func TestRoleHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(RoleHandlersTestSuite))
}
