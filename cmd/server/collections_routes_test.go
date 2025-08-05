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

// Test suite for collection handlers
type CollectionHandlersTestSuite struct {
	suite.Suite
	handler *CollectionsHandler
	db      *sql.DB
	mock    sqlmock.Sqlmock
	router  *gin.Engine
}

// SetupSuite runs once before all tests
func (suite *CollectionHandlersTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	logrus.SetLevel(logrus.PanicLevel)
}

// SetupTest runs before each test
func (suite *CollectionHandlersTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = db
	suite.mock = mock
	suite.router = gin.New()

	// Create mock server interface
	mockServer := &mockCollectionServerInterface{db: db}
	suite.handler = NewCollectionsHandler(mockServer)

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	suite.handler.SetupRoutes(v1)
}

// TearDownTest runs after each test
func (suite *CollectionHandlersTestSuite) TearDownTest() {
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
	suite.db.Close()
}

// Mock server interface for testing
type mockCollectionServerInterface struct {
	db             *sql.DB
	customAuthFunc gin.HandlerFunc
}

func (m *mockCollectionServerInterface) GetDB() *sql.DB {
	return m.db
}

func (m *mockCollectionServerInterface) AuthMiddleware() gin.HandlerFunc {
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

func (m *mockCollectionServerInterface) OptionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Status(http.StatusOK)
	}
}

// Helper function to create authenticated request
func (suite *CollectionHandlersTestSuite) createAuthenticatedRequest(method, url string, body interface{}, userID, role string) (*http.Request, *gin.Engine) {
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
	mockServer := &mockCollectionServerInterface{
		db:             suite.db,
		customAuthFunc: authMiddleware,
	}

	handler := NewCollectionsHandler(mockServer)

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

// Test GetCollections endpoint
func (suite *CollectionHandlersTestSuite) TestGetCollections_Success() {
	// Mock the query for collections
	rows := sqlmock.NewRows([]string{
		"collection", "icon", "note", "display_template", "hidden", "singleton",
		"translations", "archive_field", "archive_app_filter", "archive_value",
		"unarchive_value", "sort_field", "accountability", "color",
		"item_duplication_fields", "sort", "group", "collapse", "preview_url",
		"versioning", "created_at", "updated_at",
	}).AddRow(
		"test_collection", "folder", "Test collection", nil, false, false,
		nil, nil, true, nil, nil, nil, "all", "#6644FF",
		nil, 1, nil, "open", nil, false,
		time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT collection, icon, note").WillReturnRows(rows)

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	suite.mock.ExpectQuery("SELECT COUNT").WillReturnRows(countRows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/collections", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
	assert.Contains(suite.T(), response, "meta")
}

// Test CreateCollection endpoint
func (suite *CollectionHandlersTestSuite) TestCreateCollection_Success() {
	collectionData := CreateCollectionRequest{
		Collection: "new_collection",
		Icon:       &[]string{"folder"}[0],
		Note:       &[]string{"A new test collection"}[0],
	}

	// Mock check for existing collection
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	// Mock transaction
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("INSERT INTO collections").WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	suite.mock.ExpectCommit()

	// Mock fetching created collection
	rows := sqlmock.NewRows([]string{
		"collection", "icon", "note", "display_template", "hidden", "singleton",
		"translations", "archive_field", "archive_app_filter", "archive_value",
		"unarchive_value", "sort_field", "accountability", "color",
		"item_duplication_fields", "sort", "group", "collapse", "preview_url",
		"versioning", "created_at", "updated_at",
	}).AddRow(
		"new_collection", "folder", "A new test collection", nil, false, false,
		nil, nil, true, nil, nil, nil, "all", nil,
		nil, nil, nil, "open", nil, false,
		time.Now(), time.Now(),
	)
	suite.mock.ExpectQuery("SELECT collection, icon, note").WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/collections", collectionData, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
}

func (suite *CollectionHandlersTestSuite) TestCreateCollection_AsNonAdmin() {
	collectionData := CreateCollectionRequest{
		Collection: "new_collection",
	}

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/collections", collectionData, "test-user", "User")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Admin access required", response["error"])
}

func (suite *CollectionHandlersTestSuite) TestCreateCollection_AlreadyExists() {
	collectionData := CreateCollectionRequest{
		Collection: "existing_collection",
	}

	// Mock check for existing collection
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	suite.mock.ExpectQuery("SELECT EXISTS").WillReturnRows(existsRows)

	req, router := suite.createAuthenticatedRequest("POST", "/api/v1/collections", collectionData, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Collection already exists", response["error"])
}

// Test GetCollection endpoint
func (suite *CollectionHandlersTestSuite) TestGetCollection_Success() {
	// Mock fetching collection
	rows := sqlmock.NewRows([]string{
		"collection", "icon", "note", "display_template", "hidden", "singleton",
		"translations", "archive_field", "archive_app_filter", "archive_value",
		"unarchive_value", "sort_field", "accountability", "color",
		"item_duplication_fields", "sort", "group", "collapse", "preview_url",
		"versioning", "created_at", "updated_at",
	}).AddRow(
		"test_collection", "folder", "Test collection", nil, false, false,
		nil, nil, true, nil, nil, nil, "all", "#6644FF",
		nil, 1, nil, "open", nil, false,
		time.Now(), time.Now(),
	)
	suite.mock.ExpectQuery("SELECT collection, icon, note").WillReturnRows(rows)

	// Mock fetching fields
	fieldRows := sqlmock.NewRows([]string{
		"id", "collection", "field", "special", "interface", "options", "display",
		"display_options", "readonly", "hidden", "sort", "width", "translations",
		"note", "conditions", "required", "group", "validation", "validation_message",
		"created_at", "updated_at",
	})
	suite.mock.ExpectQuery("SELECT id, collection, field").WillReturnRows(fieldRows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/collections/test_collection", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
}

func (suite *CollectionHandlersTestSuite) TestGetCollection_NotFound() {
	// Mock collection not found
	suite.mock.ExpectQuery("SELECT collection, icon, note").WillReturnError(sql.ErrNoRows)

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/collections/nonexistent", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Collection not found", response["error"])
}

// Test UpdateCollection endpoint
func (suite *CollectionHandlersTestSuite) TestUpdateCollection_Success() {
	updateData := UpdateCollectionRequest{
		Note: &[]string{"Updated note"}[0],
		Icon: &[]string{"updated_icon"}[0],
	}

	// Mock fetching existing collection
	rows := sqlmock.NewRows([]string{
		"collection", "icon", "note", "display_template", "hidden", "singleton",
		"translations", "archive_field", "archive_app_filter", "archive_value",
		"unarchive_value", "sort_field", "accountability", "color",
		"item_duplication_fields", "sort", "group", "collapse", "preview_url",
		"versioning", "created_at", "updated_at",
	}).AddRow(
		"test_collection", "folder", "Test collection", nil, false, false,
		nil, nil, true, nil, nil, nil, "all", "#6644FF",
		nil, 1, nil, "open", nil, false,
		time.Now(), time.Now(),
	)
	suite.mock.ExpectQuery("SELECT collection, icon, note").WillReturnRows(rows)

	// Mock update
	suite.mock.ExpectExec("UPDATE collections SET").WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock fetching updated collection
	updatedRows := sqlmock.NewRows([]string{
		"collection", "icon", "note", "display_template", "hidden", "singleton",
		"translations", "archive_field", "archive_app_filter", "archive_value",
		"unarchive_value", "sort_field", "accountability", "color",
		"item_duplication_fields", "sort", "group", "collapse", "preview_url",
		"versioning", "created_at", "updated_at",
	}).AddRow(
		"test_collection", "updated_icon", "Updated note", nil, false, false,
		nil, nil, true, nil, nil, nil, "all", "#6644FF",
		nil, 1, nil, "open", nil, false,
		time.Now(), time.Now(),
	)
	suite.mock.ExpectQuery("SELECT collection, icon, note").WillReturnRows(updatedRows)

	req, router := suite.createAuthenticatedRequest("PATCH", "/api/v1/collections/test_collection", updateData, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")
}

// Test DeleteCollection endpoint
func (suite *CollectionHandlersTestSuite) TestDeleteCollection_Success() {
	// Mock fetching existing collection
	rows := sqlmock.NewRows([]string{
		"collection", "icon", "note", "display_template", "hidden", "singleton",
		"translations", "archive_field", "archive_app_filter", "archive_value",
		"unarchive_value", "sort_field", "accountability", "color",
		"item_duplication_fields", "sort", "group", "collapse", "preview_url",
		"versioning", "created_at", "updated_at",
	}).AddRow(
		"test_collection", "folder", "Test collection", nil, false, false,
		nil, nil, true, nil, nil, nil, "all", "#6644FF",
		nil, 1, nil, "open", nil, false,
		time.Now(), time.Now(),
	)
	suite.mock.ExpectQuery("SELECT collection, icon, note").WillReturnRows(rows)

	// Mock transaction for deletion
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("DELETE FROM fields WHERE collection").WillReturnResult(sqlmock.NewResult(0, 1))
	suite.mock.ExpectExec("DELETE FROM collections WHERE collection").WillReturnResult(sqlmock.NewResult(0, 1))
	suite.mock.ExpectExec("DROP TABLE IF EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	suite.mock.ExpectCommit()

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/collections/test_collection", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Collection deleted successfully", response["message"])
}

func (suite *CollectionHandlersTestSuite) TestDeleteCollection_SystemCollection() {
	// Mock fetching system collection
	rows := sqlmock.NewRows([]string{
		"collection", "icon", "note", "display_template", "hidden", "singleton",
		"translations", "archive_field", "archive_app_filter", "archive_value",
		"unarchive_value", "sort_field", "accountability", "color",
		"item_duplication_fields", "sort", "group", "collapse", "preview_url",
		"versioning", "created_at", "updated_at",
	}).AddRow(
		"users", "person", "Users collection", nil, false, false,
		nil, nil, true, nil, nil, nil, "all", "#6644FF",
		nil, 1, nil, "open", nil, false,
		time.Now(), time.Now(),
	)
	suite.mock.ExpectQuery("SELECT collection, icon, note").WillReturnRows(rows)

	req, router := suite.createAuthenticatedRequest("DELETE", "/api/v1/collections/users", nil, "test-user", "Administrator")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Cannot delete system collection", response["error"])
}

// Run the test suite
func TestCollectionHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(CollectionHandlersTestSuite))
}
