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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test suite for field handlers
type FieldHandlersTestSuite struct {
	suite.Suite
	handler *FieldsHandler
	db      *sql.DB
	mock    sqlmock.Sqlmock
	router  *gin.Engine
}

// SetupSuite runs once before all tests
func (suite *FieldHandlersTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

// SetupTest runs before each test
func (suite *FieldHandlersTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = db
	suite.mock = mock
	suite.router = gin.New()

	// Create mock server interface
	mockServer := &mockServerInterface{db: db}
	suite.handler = NewFieldsHandler(mockServer)

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	suite.handler.SetupRoutes(v1)
}

// TearDownTest runs after each test
func (suite *FieldHandlersTestSuite) TearDownTest() {
	suite.db.Close()
}

func TestFieldsHandlerSuite(t *testing.T) {
	suite.Run(t, new(FieldHandlersTestSuite))
}

var (
	testTime = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
)

func (suite *FieldHandlersTestSuite) TestCreateField() {
	suite.Run("Success", func() {
		// Mock collection existence check
		suite.mock.ExpectQuery("SELECT EXISTS").
			WithArgs("test_collection").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Mock field existence check
		suite.mock.ExpectQuery("SELECT EXISTS").
			WithArgs("test_collection", "test_field").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock transaction
		suite.mock.ExpectBegin()

		// Mock field insertion
		suite.mock.ExpectExec("INSERT INTO fields").
			WithArgs(
				"test_collection", "test_field", sqlmock.AnyArg(), "input", sqlmock.AnyArg(),
				"raw", sqlmock.AnyArg(), false, false, sqlmock.AnyArg(), "full",
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), true, sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock database column creation (ALTER TABLE)
		suite.mock.ExpectExec("ALTER TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

		suite.mock.ExpectCommit()

		// Mock field retrieval after creation
		rows := sqlmock.NewRows([]string{
			"id", "collection", "field", "special", "interface", "options", "display",
			"display_options", "readonly", "hidden", "sort", "width", "translations",
			"note", "conditions", "required", "group", "validation", "validation_message",
			"created_at", "updated_at",
		}).AddRow(
			"field-id", "test_collection", "test_field", pq.StringArray{}, "input", nil, "raw",
			nil, false, false, nil, "full", nil, nil, nil, true, nil, nil, nil,
			testTime, testTime,
		)

		suite.mock.ExpectQuery("SELECT id, collection, field").
			WithArgs("test_collection", "test_field").
			WillReturnRows(rows)

		fieldData := CreateFieldRequest{
			Field:     "test_field",
			Interface: stringPtr("input"),
			Display:   stringPtr("raw"),
			Required:  boolPtr(true),
			Hidden:    boolPtr(false),
			Width:     stringPtr("full"),
			Schema: &FieldSchema{
				DataType:   "string",
				MaxLength:  intPtr(255),
				IsNullable: boolPtr(false),
			},
		}

		body, _ := json.Marshal(fieldData)
		req, _ := http.NewRequest("POST", "/api/v1/fields/test_collection", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = addMockAuthContext(req, "admin", "Administrator")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)

		data := response["data"].(map[string]interface{})
		assert.Equal(suite.T(), "test_field", data["field"])
		assert.Equal(suite.T(), "test_collection", data["collection"])
		assert.Equal(suite.T(), "input", data["interface"])
	})

	suite.Run("DuplicateField", func() {
		// Mock collection existence check
		suite.mock.ExpectQuery("SELECT EXISTS").
			WithArgs("test_collection").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Mock field existence check (field exists)
		suite.mock.ExpectQuery("SELECT EXISTS").
			WithArgs("test_collection", "test_field").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		fieldData := CreateFieldRequest{
			Field: "test_field",
		}

		body, _ := json.Marshal(fieldData)
		req, _ := http.NewRequest("POST", "/api/v1/fields/test_collection", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = addMockAuthContext(req, "admin", "Administrator")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusConflict, w.Code)
	})

	suite.Run("NonExistentCollection", func() {
		// Mock collection existence check (collection doesn't exist)
		suite.mock.ExpectQuery("SELECT EXISTS").
			WithArgs("nonexistent").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		fieldData := CreateFieldRequest{
			Field: "test_field",
		}

		body, _ := json.Marshal(fieldData)
		req, _ := http.NewRequest("POST", "/api/v1/fields/nonexistent", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = addMockAuthContext(req, "admin", "Administrator")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})

	suite.Run("Unauthorized", func() {
		// Create a new router with custom auth middleware for this test
		customRouter := gin.New()

		// Custom auth middleware that sets user role to "User"
		customAuth := func(c *gin.Context) {
			c.Set("user_id", "550e8400-e29b-41d4-a716-446655440001")
			c.Set("user_email", "user@example.com")
			c.Set("user_role", "User")
			c.Next()
		}

		mockServer := &mockServerInterface{
			db:             suite.db,
			customAuthFunc: customAuth,
		}

		handler := NewFieldsHandler(mockServer)
		v1 := customRouter.Group("/api/v1")
		handler.SetupRoutes(v1)

		fieldData := CreateFieldRequest{
			Field: "test_field",
		}

		body, _ := json.Marshal(fieldData)
		req, _ := http.NewRequest("POST", "/api/v1/fields/test_collection", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		customRouter.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	})
}

func (suite *FieldHandlersTestSuite) TestGetFields() {
	suite.Run("AllFields", func() {
		// Mock count query
		suite.mock.ExpectQuery("SELECT COUNT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Mock fields query
		rows := sqlmock.NewRows([]string{
			"id", "collection", "field", "special", "interface", "options", "display",
			"display_options", "readonly", "hidden", "sort", "width", "translations",
			"note", "conditions", "required", "group", "validation", "validation_message",
			"created_at", "updated_at",
		}).
			AddRow("field1", "collection1", "field1", pq.StringArray{}, "input", nil, "raw",
				nil, false, false, 1, "full", nil, nil, nil, false, nil, nil, nil,
				testTime, testTime).
			AddRow("field2", "collection1", "field2", pq.StringArray{}, "textarea", nil, "raw",
				nil, false, false, 2, "full", nil, nil, nil, false, nil, nil, nil,
				testTime, testTime)

		suite.mock.ExpectQuery("SELECT id, collection, field").
			WithArgs(50, 0).
			WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/v1/fields", nil)
		req = addMockAuthContext(req, "admin", "Administrator")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)

		data := response["data"].([]interface{})
		assert.Len(suite.T(), data, 2)

		meta := response["meta"].(map[string]interface{})
		assert.Equal(suite.T(), float64(1), meta["page"])
		assert.Equal(suite.T(), float64(50), meta["limit"])
		assert.Equal(suite.T(), float64(2), meta["total"])
	})
}

func (suite *FieldHandlersTestSuite) TestGetFieldsByCollection() {
	suite.Run("Success", func() {
		// Mock collection existence check
		suite.mock.ExpectQuery("SELECT EXISTS").
			WithArgs("test_collection").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Mock fields query
		rows := sqlmock.NewRows([]string{
			"id", "collection", "field", "special", "interface", "options", "display",
			"display_options", "readonly", "hidden", "sort", "width", "translations",
			"note", "conditions", "required", "group", "validation", "validation_message",
			"created_at", "updated_at",
		}).AddRow(
			"field-id", "test_collection", "test_field", pq.StringArray{}, "input", nil, "raw",
			nil, false, false, 1, "full", nil, nil, nil, false, nil, nil, nil,
			testTime, testTime,
		)

		suite.mock.ExpectQuery("SELECT id, collection, field").
			WithArgs("test_collection").
			WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/v1/fields/test_collection", nil)
		req = addMockAuthContext(req, "admin", "Administrator")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)

		data := response["data"].([]interface{})
		assert.Len(suite.T(), data, 1)

		field := data[0].(map[string]interface{})
		assert.Equal(suite.T(), "test_field", field["field"])
		assert.Equal(suite.T(), "test_collection", field["collection"])
	})

	suite.Run("NonExistentCollection", func() {
		// Mock collection existence check (collection doesn't exist)
		suite.mock.ExpectQuery("SELECT EXISTS").
			WithArgs("nonexistent").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		req, _ := http.NewRequest("GET", "/api/v1/fields/nonexistent", nil)
		req = addMockAuthContext(req, "admin", "Administrator")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})
}

func (suite *FieldHandlersTestSuite) TestDeleteField() {
	suite.Run("Success", func() {
		// Mock field retrieval
		rows := sqlmock.NewRows([]string{
			"id", "collection", "field", "special", "interface", "options", "display",
			"display_options", "readonly", "hidden", "sort", "width", "translations",
			"note", "conditions", "required", "group", "validation", "validation_message",
			"created_at", "updated_at",
		}).AddRow(
			"field-id", "test_collection", "test_field", pq.StringArray{}, "input", nil, "raw",
			nil, false, false, 1, "full", nil, nil, nil, false, nil, nil, nil,
			testTime, testTime,
		)

		suite.mock.ExpectQuery("SELECT id, collection, field").
			WithArgs("test_collection", "test_field").
			WillReturnRows(rows)

		// Mock transaction
		suite.mock.ExpectBegin()

		// Mock field deletion
		suite.mock.ExpectExec("DELETE FROM fields").
			WithArgs("test_collection", "test_field").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock column drop
		suite.mock.ExpectExec("ALTER TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

		suite.mock.ExpectCommit()

		req, _ := http.NewRequest("DELETE", "/api/v1/fields/test_collection/test_field", nil)
		req = addMockAuthContext(req, "admin", "Administrator")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})

	suite.Run("SystemField", func() {
		// Mock field retrieval for system field check
		rows := sqlmock.NewRows([]string{
			"id", "collection", "field", "special", "interface", "options", "display",
			"display_options", "readonly", "hidden", "sort", "width", "translations",
			"note", "conditions", "required", "group", "validation", "validation_message",
			"created_at", "updated_at",
		}).AddRow(
			"sys-field-id", "test_collection", "id", pq.StringArray{}, "input", nil, "raw",
			nil, false, false, 1, "full", nil, nil, nil, true, nil, nil, nil,
			testTime, testTime,
		)

		suite.mock.ExpectQuery("SELECT id, collection, field").
			WithArgs("test_collection", "id").
			WillReturnRows(rows)

		req, _ := http.NewRequest("DELETE", "/api/v1/fields/test_collection/id", nil)
		req = addMockAuthContext(req, "admin", "Administrator")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	})
}

func (suite *FieldHandlersTestSuite) TestFieldValidation() {
	// Test valid field names
	validNames := []string{"field1", "my_field", "Field_Name", "f", "field123"}
	for _, name := range validNames {
		assert.True(suite.T(), isValidFieldName(name), "Field name %s should be valid", name)
	}

	// Test invalid field names
	invalidNames := []string{"", "1field", "field-name", "field name", "field.name", "field@name"}
	for _, name := range invalidNames {
		assert.False(suite.T(), isValidFieldName(name), "Field name %s should be invalid", name)
	}

	// Test virtual fields
	virtualInterfaces := []string{"presentation-divider", "presentation-notice", "group-raw"}
	for _, interface_ := range virtualInterfaces {
		assert.True(suite.T(), isVirtualField(&interface_), "Interface %s should be virtual", interface_)
	}

	nonVirtualInterfaces := []string{"input", "textarea", "select-dropdown"}
	for _, interface_ := range nonVirtualInterfaces {
		assert.False(suite.T(), isVirtualField(&interface_), "Interface %s should not be virtual", interface_)
	}
}

func (suite *FieldHandlersTestSuite) TestFieldSchemaOperations() {
	handler := suite.handler

	// Test data type mapping
	testCases := map[string]string{
		"string":    "VARCHAR(255)",
		"text":      "TEXT",
		"integer":   "INTEGER",
		"boolean":   "BOOLEAN",
		"timestamp": "TIMESTAMP",
		"uuid":      "UUID",
		"json":      "JSONB",
	}

	for dataType, expectedSQL := range testCases {
		result := handler.mapDataTypeToSQL(dataType, nil)
		assert.Equal(suite.T(), expectedSQL, result, "Data type %s should map to %s", dataType, expectedSQL)
	}

	// Test default value formatting
	defaultValueTests := map[interface{}]string{
		"hello":     "'hello'",
		true:        "true",
		false:       "false",
		nil:         "NULL",
		float64(42): "42",
	}

	for input, expected := range defaultValueTests {
		result := handler.formatDefaultValue(input)
		assert.Equal(suite.T(), expected, result, "Default value %v should format to %s", input, expected)
	}
}

// Helper function to add mock authentication context to a request
func addMockAuthContext(req *http.Request, userID, role string) *http.Request {
	// This is a simplified version - in real implementation,
	// we'd set proper authentication headers or context
	req.Header.Set("X-Test-User-ID", userID)
	req.Header.Set("X-Test-User-Role", role)
	return req
}

// Helper functions for pointer types
func intPtr(i int) *int {
	return &i
}
