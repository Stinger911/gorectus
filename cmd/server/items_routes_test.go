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

// Test suite for item handlers
type ItemHandlersTestSuite struct {
	suite.Suite
	handler *ItemsHandler
	db      *sql.DB
	mock    sqlmock.Sqlmock
	router  *gin.Engine
}

// SetupSuite runs once before all tests
func (suite *ItemHandlersTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	logrus.SetLevel(logrus.PanicLevel)
}

// SetupTest runs before each test
func (suite *ItemHandlersTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = db
	suite.mock = mock
	suite.router = gin.New()

	// Create mock server interface
	mockServer := &mockItemServerInterface{db: db}
	suite.handler = NewItemsHandler(mockServer)

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	suite.handler.SetupRoutes(v1)
}

// TearDownTest runs after each test
func (suite *ItemHandlersTestSuite) TearDownTest() {
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
	suite.db.Close()
}

// Mock server interface for testing
type mockItemServerInterface struct {
	db             *sql.DB
	customAuthFunc gin.HandlerFunc
}

func (m *mockItemServerInterface) GetDB() *sql.DB {
	return m.db
}

func (m *mockItemServerInterface) AuthMiddleware() gin.HandlerFunc {
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

func (m *mockItemServerInterface) OptionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Status(http.StatusOK)
	}
}

// Helper function to create authenticated request
func (suite *ItemHandlersTestSuite) createAuthenticatedRequest(method, url string, body interface{}, userID, role string) (*http.Request, *gin.Engine) {
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
	mockServer := &mockItemServerInterface{
		db:             suite.db,
		customAuthFunc: authMiddleware,
	}

	handler := NewItemsHandler(mockServer)

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

// Test GetItems endpoint
func (suite *ItemHandlersTestSuite) TestGetItems_Success() {
	// Mock collection exists check
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	// Mock the items query
	rows := sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
		AddRow("test-id-1", "Test Item 1", "Description 1", time.Now(), time.Now()).
		AddRow("test-id-2", "Test Item 2", "Description 2", time.Now(), time.Now())

	suite.mock.ExpectQuery("SELECT \\* FROM").WillReturnRows(rows)

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	suite.mock.ExpectQuery("SELECT COUNT").WillReturnRows(countRows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/items/test_collection", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
	assert.Contains(suite.T(), response, "meta")

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

func (suite *ItemHandlersTestSuite) TestGetItems_CollectionNotFound() {
	// Mock collection doesn't exist
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/items/nonexistent", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Collection not found", response["error"])
}

// Test CreateItem endpoint
func (suite *ItemHandlersTestSuite) TestCreateItem_Success() {
	itemData := Item{
		"title":       "New Test Item",
		"description": "A new test item description",
		"status":      "active",
	}

	// Mock collection exists check
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	// Mock fields query for validation
	fieldRows := sqlmock.NewRows([]string{"field", "required"}).
		AddRow("title", true).
		AddRow("description", false).
		AddRow("status", false)
	suite.mock.ExpectQuery("SELECT field, required FROM fields").WillReturnRows(fieldRows)

	// Mock insert
	suite.mock.ExpectQuery("INSERT INTO").WillReturnRows(
		sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("new-item-id", time.Now(), time.Now()),
	)

	// Mock fetching created item
	itemRows := sqlmock.NewRows([]string{"id", "title", "description", "status", "created_at", "updated_at"}).
		AddRow("new-item-id", "New Test Item", "A new test item description", "active", time.Now(), time.Now())
	suite.mock.ExpectQuery("SELECT \\* FROM").WillReturnRows(itemRows)

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/items/test_collection", itemData, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "new-item-id", data["id"])
	assert.Equal(suite.T(), "New Test Item", data["title"])
}

func (suite *ItemHandlersTestSuite) TestCreateItem_MissingRequiredField() {
	itemData := Item{
		"description": "Missing title field",
	}

	// Mock collection exists check
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	// Mock fields query for validation - title is required
	fieldRows := sqlmock.NewRows([]string{"field", "required"}).
		AddRow("title", true).
		AddRow("description", false)
	suite.mock.ExpectQuery("SELECT field, required FROM fields").WillReturnRows(fieldRows)

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/items/test_collection", itemData, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Required field 'title' is missing")
}

// Test GetItem endpoint
func (suite *ItemHandlersTestSuite) TestGetItem_Success() {
	// Mock collection exists check
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	// Mock fetching item
	itemRows := sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
		AddRow("test-item-id", "Test Item", "Test description", time.Now(), time.Now())
	suite.mock.ExpectQuery("SELECT \\* FROM").WillReturnRows(itemRows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/items/test_collection/test-item-id", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "test-item-id", data["id"])
	assert.Equal(suite.T(), "Test Item", data["title"])
}

func (suite *ItemHandlersTestSuite) TestGetItem_NotFound() {
	// Mock collection exists check
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	// Mock item not found
	suite.mock.ExpectQuery("SELECT \\* FROM").WillReturnRows(sqlmock.NewRows([]string{"id", "title"}))

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/items/test_collection/nonexistent-id", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Item not found", response["error"])
}

// Test UpdateItem endpoint
func (suite *ItemHandlersTestSuite) TestUpdateItem_Success() {
	updateData := Item{
		"title":       "Updated Title",
		"description": "Updated description",
	}

	// Mock collection exists check
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	// Mock item exists check
	itemRows := sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
		AddRow("test-item-id", "Old Title", "Old description", time.Now(), time.Now())
	suite.mock.ExpectQuery("SELECT \\* FROM").WillReturnRows(itemRows)

	// Mock update
	suite.mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock fetching updated item
	updatedRows := sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
		AddRow("test-item-id", "Updated Title", "Updated description", time.Now(), time.Now())
	suite.mock.ExpectQuery("SELECT \\* FROM").WillReturnRows(updatedRows)

	req, router := suite.createAuthenticatedRequest("PATCH", "/api/v1/items/test_collection/test-item-id", updateData, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "test-item-id", data["id"])
	assert.Equal(suite.T(), "Updated Title", data["title"])
}

// Test DeleteItem endpoint
func (suite *ItemHandlersTestSuite) TestDeleteItem_Success() {
	// Mock collection exists check
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	// Mock item exists check
	itemRows := sqlmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).
		AddRow("test-item-id", "Test Item", "Test description", time.Now(), time.Now())
	suite.mock.ExpectQuery("SELECT \\* FROM").WillReturnRows(itemRows)

	// Mock delete
	suite.mock.ExpectExec("DELETE FROM").WillReturnResult(sqlmock.NewResult(0, 1))

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/items/test_collection/test-item-id", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Item deleted successfully", response["message"])
}

// Run the test suite
func TestItemHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(ItemHandlersTestSuite))
}
