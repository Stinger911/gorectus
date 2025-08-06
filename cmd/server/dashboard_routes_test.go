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

type DashboardHandlersTestSuite struct {
	suite.Suite
	db      *sql.DB
	mock    sqlmock.Sqlmock
	router  *gin.Engine
	handler *DashboardHandler
}

// Mock server interface for testing
type mockDashboardServerInterface struct {
	db             *sql.DB
	customAuthFunc gin.HandlerFunc
}

func (m *mockDashboardServerInterface) GetDB() *sql.DB {
	return m.db
}

func (m *mockDashboardServerInterface) AuthMiddleware() gin.HandlerFunc {
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

func (m *mockDashboardServerInterface) OptionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Status(http.StatusOK)
	}
}

// SetupSuite runs once before all tests
func (suite *DashboardHandlersTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	logrus.SetLevel(logrus.PanicLevel)
}

// SetupTest runs before each test
func (suite *DashboardHandlersTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = db
	suite.mock = mock
	suite.router = gin.New()

	// Create mock server interface
	mockServer := &mockDashboardServerInterface{db: db}
	suite.handler = NewDashboardHandler(mockServer)

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	suite.handler.SetupRoutes(v1)
}

// TearDownTest runs after each test
func (suite *DashboardHandlersTestSuite) TearDownTest() {
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
	suite.db.Close()
}

// Helper function to create authenticated request
func (suite *DashboardHandlersTestSuite) createAuthenticatedRequest(method, url string, body interface{}, userID, role string) (*http.Request, *gin.Engine) {
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
	mockServer := &mockDashboardServerInterface{
		db:             suite.db,
		customAuthFunc: authMiddleware,
	}

	handler := NewDashboardHandler(mockServer)

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

// Test GetDashboardOverview endpoint
func (suite *DashboardHandlersTestSuite) mockDashboardOverviewQueries() {
	// Mock system stats queries
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE status = 'active'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM roles").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM collections").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM sessions WHERE expires > NOW\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock user insights queries
	suite.mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM users GROUP BY status").
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
			AddRow("active", 8).
			AddRow("inactive", 2))

	suite.mock.ExpectQuery("SELECT r.name, COUNT\\(u.id\\).*FROM roles r.*LEFT JOIN users u ON r.id = u.role_id.*GROUP BY r.name").
		WillReturnRows(sqlmock.NewRows([]string{"name", "count"}).
			AddRow("Administrator", 1).
			AddRow("User", 9))

	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users.*WHERE created_at >= DATE_TRUNC\\('week', NOW\\(\\)\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users.*WHERE created_at >= DATE_TRUNC\\('month', NOW\\(\\)\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	suite.mock.ExpectQuery("SELECT u.id, u.email, u.first_name, u.last_name, r.name as role,.*FROM users u.*ORDER BY u.created_at DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "status", "created_at", "last_access"}).
			AddRow("user-1", "user1@example.com", "John", "Doe", "User", "active", time.Now(), time.Now()))

	suite.mock.ExpectQuery("SELECT u.id, u.email, u.first_name, u.last_name, r.name as role,.*FROM users u.*WHERE u.last_access IS NOT NULL.*ORDER BY u.last_access DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "status", "created_at", "last_access"}).
			AddRow("user-1", "user1@example.com", "John", "Doe", "User", "active", time.Now(), time.Now()))

	// Mock collection metrics queries
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM collections").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	suite.mock.ExpectQuery("SELECT COALESCE\\(\"group\", 'Ungrouped'\\) as group_type, COUNT\\(\\*\\).*FROM collections.*GROUP BY \"group\"").
		WillReturnRows(sqlmock.NewRows([]string{"group_type", "count"}).
			AddRow("Content", 3).
			AddRow("Ungrouped", 2))

	suite.mock.ExpectQuery("SELECT collection, icon, note, hidden, singleton, created_at.*FROM collections.*ORDER BY created_at DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"collection", "icon", "note", "hidden", "singleton", "created_at"}).
			AddRow("posts", "article", "Blog posts", false, false, time.Now()))

	suite.mock.ExpectQuery("SELECT c.collection, c.icon, c.note, c.hidden, c.singleton, c.created_at,.*COUNT\\(a.id\\) as activity_count.*FROM collections c.*LEFT JOIN activity a ON c.collection = a.collection.*GROUP BY.*ORDER BY activity_count DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"collection", "icon", "note", "hidden", "singleton", "created_at", "activity_count"}).
			AddRow("posts", "article", "Blog posts", false, false, time.Now(), 15))

	// Mock recent activity query
	suite.mock.ExpectQuery("SELECT a.id, a.action, a.user_id, u.first_name, u.last_name,.*FROM activity a.*LEFT JOIN users u ON a.user_id = u.id.*ORDER BY a.timestamp DESC.*LIMIT \\$1").
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "action", "user_id", "first_name", "last_name", "collection", "item", "comment", "timestamp", "ip"}).
			AddRow("activity-1", "create", "user-1", "John", "Doe", "posts", "post-1", nil, time.Now(), "127.0.0.1"))
}

func (suite *DashboardHandlersTestSuite) TestGetDashboardOverview_AsAdmin() {
	suite.mockDashboardOverviewQueries()

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/dashboard", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "system_stats")
	assert.Contains(suite.T(), data, "user_insights")
	assert.Contains(suite.T(), data, "collection_metrics")
	assert.Contains(suite.T(), data, "recent_activity")
	assert.Contains(suite.T(), data, "system_health")
}

func (suite *DashboardHandlersTestSuite) TestGetDashboardOverview_AsNonAdmin() {
	// Mock system stats queries only (no user insights for non-admin)
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE status = 'active'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM roles").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM collections").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM sessions WHERE expires > NOW\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock collection metrics queries
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM collections").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	suite.mock.ExpectQuery("SELECT COALESCE\\(\"group\", 'Ungrouped'\\) as group_type, COUNT\\(\\*\\).*FROM collections.*GROUP BY \"group\"").
		WillReturnRows(sqlmock.NewRows([]string{"group_type", "count"}).
			AddRow("Content", 3).
			AddRow("Ungrouped", 2))

	suite.mock.ExpectQuery("SELECT collection, icon, note, hidden, singleton, created_at.*FROM collections.*ORDER BY created_at DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"collection", "icon", "note", "hidden", "singleton", "created_at"}).
			AddRow("posts", "article", "Blog posts", false, false, time.Now()))

	suite.mock.ExpectQuery("SELECT c.collection, c.icon, c.note, c.hidden, c.singleton, c.created_at,.*COUNT\\(a.id\\) as activity_count.*FROM collections c.*LEFT JOIN activity a ON c.collection = a.collection.*GROUP BY.*ORDER BY activity_count DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"collection", "icon", "note", "hidden", "singleton", "created_at", "activity_count"}).
			AddRow("posts", "article", "Blog posts", false, false, time.Now(), 15))

	// Mock recent activity query
	suite.mock.ExpectQuery("SELECT a.id, a.action, a.user_id, u.first_name, u.last_name,.*FROM activity a.*LEFT JOIN users u ON a.user_id = u.id.*ORDER BY a.timestamp DESC.*LIMIT \\$1").
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "action", "user_id", "first_name", "last_name", "collection", "item", "comment", "timestamp", "ip"}).
			AddRow("activity-1", "create", "user-1", "John", "Doe", "posts", "post-1", nil, time.Now(), "127.0.0.1"))

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/dashboard", nil, "user-id", "Public")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "system_stats")
	assert.NotContains(suite.T(), data, "user_insights") // Non-admin should not see user insights
	assert.Contains(suite.T(), data, "collection_metrics")
	assert.Contains(suite.T(), data, "recent_activity")
	assert.Contains(suite.T(), data, "system_health")
}

// Test GetSystemStats endpoint
func (suite *DashboardHandlersTestSuite) TestGetSystemStats_AsAdmin() {
	// Mock system stats queries
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE status = 'active'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM roles").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM collections").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM sessions WHERE expires > NOW\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/dashboard/stats", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(10), data["total_users"])
	assert.Equal(suite.T(), float64(8), data["active_users"])
	assert.Equal(suite.T(), float64(3), data["total_roles"])
	assert.Equal(suite.T(), float64(5), data["total_collections"])
	assert.Equal(suite.T(), float64(2), data["total_sessions"])
}

// Test GetRecentActivity endpoint
func (suite *DashboardHandlersTestSuite) TestGetRecentActivity_AsAdmin() {
	suite.mock.ExpectQuery("SELECT a.id, a.action, a.user_id, u.first_name, u.last_name,.*FROM activity a.*LEFT JOIN users u ON a.user_id = u.id.*ORDER BY a.timestamp DESC.*LIMIT \\$1").
		WithArgs(20). // Default limit
		WillReturnRows(sqlmock.NewRows([]string{"id", "action", "user_id", "first_name", "last_name", "collection", "item", "comment", "timestamp", "ip"}).
			AddRow("activity-1", "create", "user-1", "John", "Doe", "posts", "post-1", nil, time.Now(), "127.0.0.1").
			AddRow("activity-2", "update", "user-2", "Jane", "Smith", "posts", "post-2", "Updated content", time.Now(), "127.0.0.1"))

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/dashboard/activity", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

// Test GetRecentActivity with custom limit
func (suite *DashboardHandlersTestSuite) TestGetRecentActivity_WithLimit() {
	suite.mock.ExpectQuery("SELECT a.id, a.action, a.user_id, u.first_name, u.last_name,.*FROM activity a.*LEFT JOIN users u ON a.user_id = u.id.*ORDER BY a.timestamp DESC.*LIMIT \\$1").
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "action", "user_id", "first_name", "last_name", "collection", "item", "comment", "timestamp", "ip"}).
			AddRow("activity-1", "create", "user-1", "John", "Doe", "posts", "post-1", nil, time.Now(), "127.0.0.1"))

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/dashboard/activity?limit=5", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// Test GetUserInsights endpoint
func (suite *DashboardHandlersTestSuite) TestGetUserInsights_AsAdmin() {
	// Mock user insights queries
	suite.mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM users GROUP BY status").
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
			AddRow("active", 8).
			AddRow("inactive", 2))

	suite.mock.ExpectQuery("SELECT r.name, COUNT\\(u.id\\).*FROM roles r.*LEFT JOIN users u ON r.id = u.role_id.*GROUP BY r.name").
		WillReturnRows(sqlmock.NewRows([]string{"name", "count"}).
			AddRow("Administrator", 1).
			AddRow("User", 9))

	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users.*WHERE created_at >= DATE_TRUNC\\('week', NOW\\(\\)\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users.*WHERE created_at >= DATE_TRUNC\\('month', NOW\\(\\)\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	suite.mock.ExpectQuery("SELECT u.id, u.email, u.first_name, u.last_name, r.name as role,.*FROM users u.*ORDER BY u.created_at DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "status", "created_at", "last_access"}).
			AddRow("user-1", "user1@example.com", "John", "Doe", "User", "active", time.Now(), time.Now()))

	suite.mock.ExpectQuery("SELECT u.id, u.email, u.first_name, u.last_name, r.name as role,.*FROM users u.*WHERE u.last_access IS NOT NULL.*ORDER BY u.last_access DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "status", "created_at", "last_access"}).
			AddRow("user-1", "user1@example.com", "John", "Doe", "User", "active", time.Now(), time.Now()))

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/dashboard/users", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "users_by_status")
	assert.Contains(suite.T(), data, "users_by_role")
	assert.Contains(suite.T(), data, "new_users_this_week")
	assert.Contains(suite.T(), data, "new_users_this_month")
	assert.Contains(suite.T(), data, "recent_registrations")
	assert.Contains(suite.T(), data, "most_active_users")
}

// Test GetUserInsights endpoint as non-admin (should return 403)
func (suite *DashboardHandlersTestSuite) TestGetUserInsights_AsNonAdmin() {
	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/dashboard/users", nil, "user-id", "Public")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "error")
	assert.Equal(suite.T(), "Admin access required", response["error"])
}

// Test GetCollectionInsights endpoint
func (suite *DashboardHandlersTestSuite) TestGetCollectionInsights_AsAdmin() {
	// Mock collection metrics queries
	suite.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM collections").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	suite.mock.ExpectQuery("SELECT COALESCE\\(\"group\", 'Ungrouped'\\) as group_type, COUNT\\(\\*\\).*FROM collections.*GROUP BY \"group\"").
		WillReturnRows(sqlmock.NewRows([]string{"group_type", "count"}).
			AddRow("Content", 3).
			AddRow("Ungrouped", 2))

	suite.mock.ExpectQuery("SELECT collection, icon, note, hidden, singleton, created_at.*FROM collections.*ORDER BY created_at DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"collection", "icon", "note", "hidden", "singleton", "created_at"}).
			AddRow("posts", "article", "Blog posts", false, false, time.Now()))

	suite.mock.ExpectQuery("SELECT c.collection, c.icon, c.note, c.hidden, c.singleton, c.created_at,.*COUNT\\(a.id\\) as activity_count.*FROM collections c.*LEFT JOIN activity a ON c.collection = a.collection.*GROUP BY.*ORDER BY activity_count DESC.*LIMIT 10").
		WillReturnRows(sqlmock.NewRows([]string{"collection", "icon", "note", "hidden", "singleton", "created_at", "activity_count"}).
			AddRow("posts", "article", "Blog posts", false, false, time.Now(), 15))

	req, router := suite.createAuthenticatedRequest("GET", "/api/v1/dashboard/collections", nil, "admin-id", "Administrator")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "data")

	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "total_collections")
	assert.Contains(suite.T(), data, "collections_by_type")
	assert.Contains(suite.T(), data, "recent_collections")
	assert.Contains(suite.T(), data, "most_active_collections")
}

// Run the test suite
func TestDashboardHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(DashboardHandlersTestSuite))
}
