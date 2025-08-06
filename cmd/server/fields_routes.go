package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// FieldsHandler handles field-related routes
type FieldsHandler struct {
	db             *sql.DB
	authMiddleware gin.HandlerFunc
	optionsHandler gin.HandlerFunc
}

// NewFieldsHandler creates a new fields handler
func NewFieldsHandler(server ServerInterface) *FieldsHandler {
	return &FieldsHandler{
		db:             server.GetDB(),
		authMiddleware: server.AuthMiddleware(),
		optionsHandler: server.OptionsHandler(),
	}
}

// SetupRoutes sets up field routes
func (h *FieldsHandler) SetupRoutes(v1 *gin.RouterGroup) {
	// CORS preflight OPTIONS for fields endpoints
	v1.OPTIONS("/fields", h.optionsHandler)
	v1.OPTIONS("/fields/:collection", h.optionsHandler)
	v1.OPTIONS("/fields/:collection/:field", h.optionsHandler)

	// Fields routes (protected)
	fields := v1.Group("/fields")
	fields.Use(h.authMiddleware)
	{
		// Get all fields or fields for specific collection
		fields.GET("", h.getFields)
		fields.GET("/:collection", h.getFieldsByCollection)

		// Individual field operations
		fields.GET("/:collection/:field", h.getField)
		fields.POST("/:collection", h.createField)
		fields.PATCH("/:collection/:field", h.updateField)
		fields.DELETE("/:collection/:field", h.deleteField)
	}
}

// FieldDetail represents a complete field definition
type FieldDetail struct {
	ID                string      `json:"id"`
	Collection        string      `json:"collection"`
	Field             string      `json:"field"`
	Special           []string    `json:"special"`
	Interface         *string     `json:"interface"`
	Options           interface{} `json:"options"`
	Display           *string     `json:"display"`
	DisplayOptions    interface{} `json:"display_options"`
	Readonly          bool        `json:"readonly"`
	Hidden            bool        `json:"hidden"`
	Sort              *int        `json:"sort"`
	Width             string      `json:"width"`
	Translations      interface{} `json:"translations"`
	Note              *string     `json:"note"`
	Conditions        interface{} `json:"conditions"`
	Required          bool        `json:"required"`
	Group             *string     `json:"group"`
	Validation        interface{} `json:"validation"`
	ValidationMessage *string     `json:"validation_message"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

// Request types for fields
type CreateFieldRequest struct {
	Field             string       `json:"field" binding:"required"`
	Special           []string     `json:"special"`
	Interface         *string      `json:"interface"`
	Options           interface{}  `json:"options"`
	Display           *string      `json:"display"`
	DisplayOptions    interface{}  `json:"display_options"`
	Readonly          *bool        `json:"readonly"`
	Hidden            *bool        `json:"hidden"`
	Sort              *int         `json:"sort"`
	Width             *string      `json:"width"`
	Translations      interface{}  `json:"translations"`
	Note              *string      `json:"note"`
	Conditions        interface{}  `json:"conditions"`
	Required          *bool        `json:"required"`
	Group             *string      `json:"group"`
	Validation        interface{}  `json:"validation"`
	ValidationMessage *string      `json:"validation_message"`
	Schema            *FieldSchema `json:"schema"` // For creating database columns
}

type UpdateFieldRequest struct {
	Special           []string     `json:"special"`
	Interface         *string      `json:"interface"`
	Options           interface{}  `json:"options"`
	Display           *string      `json:"display"`
	DisplayOptions    interface{}  `json:"display_options"`
	Readonly          *bool        `json:"readonly"`
	Hidden            *bool        `json:"hidden"`
	Sort              *int         `json:"sort"`
	Width             *string      `json:"width"`
	Translations      interface{}  `json:"translations"`
	Note              *string      `json:"note"`
	Conditions        interface{}  `json:"conditions"`
	Required          *bool        `json:"required"`
	Group             *string      `json:"group"`
	Validation        interface{}  `json:"validation"`
	ValidationMessage *string      `json:"validation_message"`
	Schema            *FieldSchema `json:"schema"` // For altering database columns
}

// FieldSchema represents database column schema information
type FieldSchema struct {
	DataType      string      `json:"data_type"`      // varchar, integer, boolean, text, json, uuid, timestamp, etc.
	MaxLength     *int        `json:"max_length"`     // For varchar
	IsNullable    *bool       `json:"is_nullable"`    // Whether column allows NULL
	DefaultValue  interface{} `json:"default_value"`  // Default value for column
	IsUnique      *bool       `json:"is_unique"`      // Whether column should be unique
	IsPrimaryKey  *bool       `json:"is_primary_key"` // Whether column is primary key
	ForeignTable  *string     `json:"foreign_table"`  // For foreign key relationships
	ForeignColumn *string     `json:"foreign_column"` // For foreign key relationships
}

// FieldsListResponse represents a paginated list of fields
type FieldsListResponse struct {
	Data []FieldDetail `json:"data"`
	Meta struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"meta"`
}

// isAdmin checks if the requesting user is an admin
func (h *FieldsHandler) isAdmin(c *gin.Context) bool {
	currentUserRole := c.GetString("user_role")
	return currentUserRole == "Administrator"
}

// getFields returns all fields in the system with optional filtering
//
//	@Summary		Get all fields
//	@Description	Retrieve a paginated list of all fields in the system with optional collection filtering
//	@Tags			fields
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number for pagination (default: 1)"
//	@Param			limit		query		int		false	"Number of items per page (max: 100, default: 50)"
//	@Param			collection	query		string	false	"Filter fields by collection name"
//	@Success		200			{object}	FieldsListResponse	"List of fields with pagination metadata"
//	@Failure		401			{object}	ErrorResponse		"Unauthorized"
//	@Failure		500			{object}	ErrorResponse		"Internal server error"
//	@Router			/fields [get]
func (h *FieldsHandler) getFields(c *gin.Context) {
	// Parse query parameters
	page := 1
	limit := 50
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	collection := c.Query("collection")
	offset := (page - 1) * limit

	// Build query with optional collection filter
	baseQuery := `
		SELECT id, collection, field, special, interface, options, display, 
		       display_options, readonly, hidden, sort, width, translations,
		       note, conditions, required, "group", validation, validation_message,
		       created_at, updated_at
		FROM fields`

	countQuery := "SELECT COUNT(*) FROM fields"
	args := []interface{}{}
	whereClause := ""

	if collection != "" {
		whereClause = " WHERE collection = $1"
		args = append(args, collection)
		countQuery += whereClause
		baseQuery += whereClause
	}

	// Add ordering and pagination
	baseQuery += " ORDER BY collection ASC, sort ASC, field ASC LIMIT $" +
		strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	// Get total count
	var total int
	countArgs := args
	if collection != "" {
		countArgs = args[:1] // Only collection parameter for count
	}
	err := h.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Error counting fields")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Get fields
	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		logrus.WithError(err).Error("Database error while fetching fields")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var fields []FieldDetail
	for rows.Next() {
		field, err := h.scanFieldRow(rows)
		if err != nil {
			logrus.WithError(err).Error("Error scanning field row")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		fields = append(fields, *field)
	}

	response := FieldsListResponse{
		Data: fields,
	}
	response.Meta.Page = page
	response.Meta.Limit = limit
	response.Meta.Total = total

	c.JSON(http.StatusOK, response)
}

// getFieldsByCollection returns all fields for a specific collection
//
//	@Summary		Get fields by collection
//	@Description	Retrieve all fields for a specific collection
//	@Tags			fields
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string	true	"Collection name"
//	@Success		200			{object}	map[string][]FieldDetail	"List of fields for the collection"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		404			{object}	ErrorResponse	"Collection not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/fields/{collection} [get]
func (h *FieldsHandler) getFieldsByCollection(c *gin.Context) {
	collectionName := c.Param("collection")

	// Check if collection exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM collections WHERE collection = $1)", collectionName).Scan(&exists)
	if err != nil {
		logrus.WithError(err).Error("Database error while checking collection existence")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	fields, err := h.getFieldsByCollectionName(collectionName)
	if err != nil {
		logrus.WithError(err).Error("Database error while fetching collection fields")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": fields})
}

// getField returns a specific field
//
//	@Summary		Get field by name
//	@Description	Retrieve a specific field by collection and field name
//	@Tags			fields
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string	true	"Collection name"
//	@Param			field		path		string	true	"Field name"
//	@Success		200			{object}	map[string]FieldDetail	"Field details"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		404			{object}	ErrorResponse	"Field not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/fields/{collection}/{field} [get]
func (h *FieldsHandler) getField(c *gin.Context) {
	collectionName := c.Param("collection")
	fieldName := c.Param("field")

	field, err := h.getFieldByName(collectionName, fieldName)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching field")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": field})
}

// createField creates a new field in a collection
//
//	@Summary		Create a new field
//	@Description	Create a new field in a collection with optional database column creation
//	@Tags			fields
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string				true	"Collection name"
//	@Param			field		body		CreateFieldRequest	true	"Field creation data"
//	@Success		201			{object}	map[string]FieldDetail	"Created field details"
//	@Failure		400			{object}	ErrorResponse	"Bad request (invalid payload, field name, or field already exists)"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		403			{object}	ErrorResponse	"Forbidden (admin access required)"
//	@Failure		404			{object}	ErrorResponse	"Collection not found"
//	@Failure		409			{object}	ErrorResponse	"Field already exists"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/fields/{collection} [post]
func (h *FieldsHandler) createField(c *gin.Context) {
	// Only admins can create fields
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	collectionName := c.Param("collection")

	// Check if collection exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM collections WHERE collection = $1)", collectionName).Scan(&exists)
	if err != nil {
		logrus.WithError(err).Error("Database error while checking collection existence")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	var req CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid create field request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if field already exists
	var fieldExists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM fields WHERE collection = $1 AND field = $2)",
		collectionName, req.Field).Scan(&fieldExists)
	if err != nil {
		logrus.WithError(err).Error("Database error while checking field existence")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if fieldExists {
		c.JSON(http.StatusConflict, gin.H{"error": "Field already exists"})
		return
	}

	// Validate field name (no spaces, special characters except underscore)
	if !isValidFieldName(req.Field) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid field name. Use only letters, numbers, and underscores"})
		return
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		logrus.WithError(err).Error("Failed to begin transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Set defaults
	readonly := false
	hidden := false
	required := false
	width := "full"

	if req.Readonly != nil {
		readonly = *req.Readonly
	}
	if req.Hidden != nil {
		hidden = *req.Hidden
	}
	if req.Required != nil {
		required = *req.Required
	}
	if req.Width != nil {
		width = *req.Width
	}

	// Handle JSON fields
	var optionsData, displayOptionsData, translationsData, conditionsData, validationData interface{}

	if req.Options != nil {
		optionsBytes, _ := json.Marshal(req.Options)
		optionsData = optionsBytes
	}
	if req.DisplayOptions != nil {
		displayOptionsBytes, _ := json.Marshal(req.DisplayOptions)
		displayOptionsData = displayOptionsBytes
	}
	if req.Translations != nil {
		translationsBytes, _ := json.Marshal(req.Translations)
		translationsData = translationsBytes
	}
	if req.Conditions != nil {
		conditionsBytes, _ := json.Marshal(req.Conditions)
		conditionsData = conditionsBytes
	}
	if req.Validation != nil {
		validationBytes, _ := json.Marshal(req.Validation)
		validationData = validationBytes
	}

	// Convert special array to PostgreSQL array format
	var special interface{}
	if len(req.Special) > 0 {
		special = pq.Array(req.Special)
	} else {
		special = pq.Array([]string{})
	}

	// Insert field metadata
	_, err = tx.Exec(`
		INSERT INTO fields (
			collection, field, special, interface, options, display, 
			display_options, readonly, hidden, sort, width, translations,
			note, conditions, required, "group", validation, validation_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, collectionName, req.Field, special, req.Interface, optionsData,
		req.Display, displayOptionsData, readonly, hidden,
		req.Sort, width, translationsData, req.Note, conditionsData,
		required, req.Group, validationData, req.ValidationMessage)

	if err != nil {
		logrus.WithError(err).Error("Database error while creating field")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Create database column if schema is provided and field is not virtual
	if req.Schema != nil && !isVirtualField(req.Interface) {
		err = h.createDatabaseColumn(tx, collectionName, req.Field, req.Schema)
		if err != nil {
			logrus.WithError(err).Error("Failed to create database column")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create database column"})
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		logrus.WithError(err).Error("Failed to commit transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Return the created field
	field, err := h.getFieldByName(collectionName, req.Field)
	if err != nil {
		logrus.WithError(err).Error("Error fetching created field")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"collection": collectionName,
		"field":      req.Field,
	}).Info("Field created successfully")
	c.JSON(http.StatusCreated, gin.H{"data": field})
}

// updateField updates an existing field
//
//	@Summary		Update an existing field
//	@Description	Update an existing field's metadata and optionally alter the database column
//	@Tags			fields
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string				true	"Collection name"
//	@Param			field		path		string				true	"Field name"
//	@Param			updates		body		UpdateFieldRequest	true	"Field update data"
//	@Success		200			{object}	map[string]FieldDetail	"Updated field details"
//	@Failure		400			{object}	ErrorResponse	"Bad request (invalid payload or no fields to update)"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		403			{object}	ErrorResponse	"Forbidden (admin access required)"
//	@Failure		404			{object}	ErrorResponse	"Field not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/fields/{collection}/{field} [patch]
func (h *FieldsHandler) updateField(c *gin.Context) {
	// Only admins can update fields
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	collectionName := c.Param("collection")
	fieldName := c.Param("field")

	// Check if field exists
	_, err := h.getFieldByName(collectionName, fieldName)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching field")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var req UpdateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid update field request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		logrus.WithError(err).Error("Failed to begin transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Build update query dynamically
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Special != nil {
		updateFields = append(updateFields, "special = $"+strconv.Itoa(argIndex))
		args = append(args, pq.Array(req.Special))
		argIndex++
	}
	if req.Interface != nil {
		updateFields = append(updateFields, "interface = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Interface)
		argIndex++
	}
	if req.Options != nil {
		optionsBytes, _ := json.Marshal(req.Options)
		updateFields = append(updateFields, "options = $"+strconv.Itoa(argIndex))
		args = append(args, optionsBytes)
		argIndex++
	}
	if req.Display != nil {
		updateFields = append(updateFields, "display = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Display)
		argIndex++
	}
	if req.DisplayOptions != nil {
		displayOptionsBytes, _ := json.Marshal(req.DisplayOptions)
		updateFields = append(updateFields, "display_options = $"+strconv.Itoa(argIndex))
		args = append(args, displayOptionsBytes)
		argIndex++
	}
	if req.Readonly != nil {
		updateFields = append(updateFields, "readonly = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Readonly)
		argIndex++
	}
	if req.Hidden != nil {
		updateFields = append(updateFields, "hidden = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Hidden)
		argIndex++
	}
	if req.Sort != nil {
		updateFields = append(updateFields, "sort = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Sort)
		argIndex++
	}
	if req.Width != nil {
		updateFields = append(updateFields, "width = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Width)
		argIndex++
	}
	if req.Translations != nil {
		translationsBytes, _ := json.Marshal(req.Translations)
		updateFields = append(updateFields, "translations = $"+strconv.Itoa(argIndex))
		args = append(args, translationsBytes)
		argIndex++
	}
	if req.Note != nil {
		updateFields = append(updateFields, "note = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Note)
		argIndex++
	}
	if req.Conditions != nil {
		conditionsBytes, _ := json.Marshal(req.Conditions)
		updateFields = append(updateFields, "conditions = $"+strconv.Itoa(argIndex))
		args = append(args, conditionsBytes)
		argIndex++
	}
	if req.Required != nil {
		updateFields = append(updateFields, "required = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Required)
		argIndex++
	}
	if req.Group != nil {
		updateFields = append(updateFields, "\"group\" = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Group)
		argIndex++
	}
	if req.Validation != nil {
		validationBytes, _ := json.Marshal(req.Validation)
		updateFields = append(updateFields, "validation = $"+strconv.Itoa(argIndex))
		args = append(args, validationBytes)
		argIndex++
	}
	if req.ValidationMessage != nil {
		updateFields = append(updateFields, "validation_message = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ValidationMessage)
		argIndex++
	}

	if len(updateFields) == 0 && req.Schema == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Update field metadata if there are changes
	if len(updateFields) > 0 {
		updateFields = append(updateFields, "updated_at = CURRENT_TIMESTAMP")
		args = append(args, collectionName, fieldName)

		query := "UPDATE fields SET " + strings.Join(updateFields, ", ") +
			" WHERE collection = $" + strconv.Itoa(argIndex) + " AND field = $" + strconv.Itoa(argIndex+1)

		_, err = tx.Exec(query, args...)
		if err != nil {
			logrus.WithError(err).Error("Database error while updating field")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	}

	// Update database column if schema changes are provided
	if req.Schema != nil && !isVirtualField(req.Interface) {
		err = h.alterDatabaseColumn(tx, collectionName, fieldName, req.Schema)
		if err != nil {
			logrus.WithError(err).Error("Failed to alter database column")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to alter database column"})
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		logrus.WithError(err).Error("Failed to commit transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Return updated field
	field, err := h.getFieldByName(collectionName, fieldName)
	if err != nil {
		logrus.WithError(err).Error("Error fetching updated field")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"collection": collectionName,
		"field":      fieldName,
	}).Info("Field updated successfully")
	c.JSON(http.StatusOK, gin.H{"data": field})
}

// deleteField deletes a field from a collection
//
//	@Summary		Delete a field
//	@Description	Delete a field from a collection and optionally drop the database column
//	@Tags			fields
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string	true	"Collection name"
//	@Param			field		path		string	true	"Field name"
//	@Success		200			{object}	SuccessMessage	"Field deleted successfully"
//	@Failure		400			{object}	ErrorResponse	"Bad request (cannot delete system field)"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		403			{object}	ErrorResponse	"Forbidden (admin access required)"
//	@Failure		404			{object}	ErrorResponse	"Field not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/fields/{collection}/{field} [delete]
func (h *FieldsHandler) deleteField(c *gin.Context) {
	// Only admins can delete fields
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	collectionName := c.Param("collection")
	fieldName := c.Param("field")

	// Check if field exists
	field, err := h.getFieldByName(collectionName, fieldName)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching field")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Prevent deletion of system fields
	systemFields := []string{"id", "created_at", "updated_at"}
	for _, sysField := range systemFields {
		if fieldName == sysField {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete system field"})
			return
		}
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		logrus.WithError(err).Error("Failed to begin transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Delete field metadata
	_, err = tx.Exec("DELETE FROM fields WHERE collection = $1 AND field = $2", collectionName, fieldName)
	if err != nil {
		logrus.WithError(err).Error("Database error while deleting field")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Drop database column if it's not a virtual field
	if !isVirtualField(field.Interface) {
		err = h.dropDatabaseColumn(tx, collectionName, fieldName)
		if err != nil {
			logrus.WithError(err).Error("Failed to drop database column")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to drop database column"})
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		logrus.WithError(err).Error("Failed to commit transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"collection": collectionName,
		"field":      fieldName,
	}).Info("Field deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Field deleted successfully"})
}

// Helper methods

// scanFieldRow scans a database row into a FieldDetail struct
func (h *FieldsHandler) scanFieldRow(rows *sql.Rows) (*FieldDetail, error) {
	var field FieldDetail
	var special pq.StringArray
	var optionsBytes, displayOptionsBytes, translationsBytes, conditionsBytes, validationBytes []byte

	err := rows.Scan(
		&field.ID, &field.Collection, &field.Field, &special, &field.Interface,
		&optionsBytes, &field.Display, &displayOptionsBytes, &field.Readonly,
		&field.Hidden, &field.Sort, &field.Width, &translationsBytes, &field.Note,
		&conditionsBytes, &field.Required, &field.Group, &validationBytes,
		&field.ValidationMessage, &field.CreatedAt, &field.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert PostgreSQL array to Go slice
	field.Special = []string(special)

	// Parse JSON fields
	if optionsBytes != nil {
		json.Unmarshal(optionsBytes, &field.Options)
	}
	if displayOptionsBytes != nil {
		json.Unmarshal(displayOptionsBytes, &field.DisplayOptions)
	}
	if translationsBytes != nil {
		json.Unmarshal(translationsBytes, &field.Translations)
	}
	if conditionsBytes != nil {
		json.Unmarshal(conditionsBytes, &field.Conditions)
	}
	if validationBytes != nil {
		json.Unmarshal(validationBytes, &field.Validation)
	}

	return &field, nil
}

// getFieldsByCollectionName returns all fields for a collection
func (h *FieldsHandler) getFieldsByCollectionName(collectionName string) ([]FieldDetail, error) {
	query := `
		SELECT id, collection, field, special, interface, options, display, 
		       display_options, readonly, hidden, sort, width, translations,
		       note, conditions, required, "group", validation, validation_message,
		       created_at, updated_at
		FROM fields 
		WHERE collection = $1
		ORDER BY sort ASC, field ASC
	`

	rows, err := h.db.Query(query, collectionName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []FieldDetail
	for rows.Next() {
		field, err := h.scanFieldRow(rows)
		if err != nil {
			return nil, err
		}
		fields = append(fields, *field)
	}

	return fields, nil
}

// getFieldByName returns a specific field by collection and field name
func (h *FieldsHandler) getFieldByName(collectionName, fieldName string) (*FieldDetail, error) {
	query := `
		SELECT id, collection, field, special, interface, options, display, 
		       display_options, readonly, hidden, sort, width, translations,
		       note, conditions, required, "group", validation, validation_message,
		       created_at, updated_at
		FROM fields 
		WHERE collection = $1 AND field = $2
	`

	rows, err := h.db.Query(query, collectionName, fieldName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	return h.scanFieldRow(rows)
}

// isValidFieldName checks if a field name is valid (letters, numbers, underscores only)
func isValidFieldName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}

	// Ensure it doesn't start with a number
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}

	return true
}

// isVirtualField checks if a field interface type represents a virtual field (doesn't need database column)
func isVirtualField(interfaceType *string) bool {
	if interfaceType == nil {
		return false
	}

	virtualInterfaces := []string{
		"presentation-divider",
		"presentation-notice",
		"group-raw",
		"group-detail",
		"alias",
	}

	for _, virtual := range virtualInterfaces {
		if *interfaceType == virtual {
			return true
		}
	}

	return false
}

// createDatabaseColumn creates a new column in the collection table
func (h *FieldsHandler) createDatabaseColumn(tx *sql.Tx, collectionName, fieldName string, schema *FieldSchema) error {
	// Build ALTER TABLE statement
	columnDef := h.buildColumnDefinition(fieldName, schema)

	alterSQL := `ALTER TABLE "` + collectionName + `" ADD COLUMN ` + columnDef

	_, err := tx.Exec(alterSQL)
	if err != nil {
		return err
	}

	// Add unique constraint if needed
	if schema.IsUnique != nil && *schema.IsUnique {
		indexSQL := `CREATE UNIQUE INDEX "idx_` + collectionName + `_` + fieldName + `_unique" ON "` +
			collectionName + `" ("` + fieldName + `")`
		_, err = tx.Exec(indexSQL)
		if err != nil {
			return err
		}
	}

	// Add foreign key constraint if needed
	if schema.ForeignTable != nil && schema.ForeignColumn != nil {
		constraintSQL := `ALTER TABLE "` + collectionName + `" ADD CONSTRAINT "fk_` + collectionName +
			`_` + fieldName + `" FOREIGN KEY ("` + fieldName + `") REFERENCES "` +
			*schema.ForeignTable + `" ("` + *schema.ForeignColumn + `")`
		_, err = tx.Exec(constraintSQL)
		if err != nil {
			return err
		}
	}

	return nil
}

// alterDatabaseColumn alters an existing column in the collection table
func (h *FieldsHandler) alterDatabaseColumn(tx *sql.Tx, collectionName, fieldName string, schema *FieldSchema) error {
	// Note: PostgreSQL has limitations on altering columns.
	// This is a simplified implementation - in production, you might need more sophisticated logic

	if schema.DataType != "" {
		alterSQL := `ALTER TABLE "` + collectionName + `" ALTER COLUMN "` + fieldName +
			`" TYPE ` + h.mapDataTypeToSQL(schema.DataType, schema.MaxLength)
		_, err := tx.Exec(alterSQL)
		if err != nil {
			return err
		}
	}

	if schema.IsNullable != nil {
		nullClause := "DROP NOT NULL"
		if !*schema.IsNullable {
			nullClause = "SET NOT NULL"
		}
		alterSQL := `ALTER TABLE "` + collectionName + `" ALTER COLUMN "` + fieldName + `" ` + nullClause
		_, err := tx.Exec(alterSQL)
		if err != nil {
			return err
		}
	}

	if schema.DefaultValue != nil {
		defaultVal := h.formatDefaultValue(schema.DefaultValue)
		alterSQL := `ALTER TABLE "` + collectionName + `" ALTER COLUMN "` + fieldName +
			`" SET DEFAULT ` + defaultVal
		_, err := tx.Exec(alterSQL)
		if err != nil {
			return err
		}
	}

	return nil
}

// dropDatabaseColumn drops a column from the collection table
func (h *FieldsHandler) dropDatabaseColumn(tx *sql.Tx, collectionName, fieldName string) error {
	dropSQL := `ALTER TABLE "` + collectionName + `" DROP COLUMN "` + fieldName + `" CASCADE`
	_, err := tx.Exec(dropSQL)
	return err
}

// buildColumnDefinition builds a SQL column definition string
func (h *FieldsHandler) buildColumnDefinition(fieldName string, schema *FieldSchema) string {
	def := `"` + fieldName + `" ` + h.mapDataTypeToSQL(schema.DataType, schema.MaxLength)

	if schema.IsNullable != nil && !*schema.IsNullable {
		def += " NOT NULL"
	}

	if schema.DefaultValue != nil {
		def += " DEFAULT " + h.formatDefaultValue(schema.DefaultValue)
	}

	return def
}

// mapDataTypeToSQL maps field data types to PostgreSQL types
func (h *FieldsHandler) mapDataTypeToSQL(dataType string, maxLength *int) string {
	switch strings.ToLower(dataType) {
	case "string", "varchar":
		if maxLength != nil && *maxLength > 0 {
			return "VARCHAR(" + strconv.Itoa(*maxLength) + ")"
		}
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "integer", "int":
		return "INTEGER"
	case "bigint":
		return "BIGINT"
	case "float", "decimal":
		return "DECIMAL"
	case "boolean", "bool":
		return "BOOLEAN"
	case "date":
		return "DATE"
	case "time":
		return "TIME"
	case "datetime", "timestamp":
		return "TIMESTAMP"
	case "uuid":
		return "UUID"
	case "json", "jsonb":
		return "JSONB"
	default:
		return "TEXT" // fallback
	}
}

// formatDefaultValue formats a default value for SQL
func (h *FieldsHandler) formatDefaultValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "NULL"
	default:
		return strconv.Itoa(int(v.(float64)))
	}
}
