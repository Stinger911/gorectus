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

// CollectionsHandler handles collection-related routes
type CollectionsHandler struct {
	db             *sql.DB
	authMiddleware gin.HandlerFunc
	optionsHandler gin.HandlerFunc
}

// NewCollectionsHandler creates a new collections handler
func NewCollectionsHandler(server ServerInterface) *CollectionsHandler {
	return &CollectionsHandler{
		db:             server.GetDB(),
		authMiddleware: server.AuthMiddleware(),
		optionsHandler: server.OptionsHandler(),
	}
}

// SetupRoutes sets up collection routes
func (h *CollectionsHandler) SetupRoutes(v1 *gin.RouterGroup) {
	// CORS preflight OPTIONS for collections endpoints
	v1.OPTIONS("/collections", h.optionsHandler)
	v1.OPTIONS("/collections/:collection", h.optionsHandler)

	// Collections routes (protected)
	collections := v1.Group("/collections")
	collections.Use(h.authMiddleware)
	{
		collections.GET("", h.getCollections)
		collections.POST("", h.createCollection)
		collections.GET("/:collection", h.getCollection)
		collections.PATCH("/:collection", h.updateCollection)
		collections.DELETE("/:collection", h.deleteCollection)
	}
}

// Collection represents a collection in the system
type Collection struct {
	Collection            string      `json:"collection"`
	Icon                  *string     `json:"icon"`
	Note                  *string     `json:"note"`
	DisplayTemplate       *string     `json:"display_template"`
	Hidden                bool        `json:"hidden"`
	Singleton             bool        `json:"singleton"`
	Translations          interface{} `json:"translations"`
	ArchiveField          *string     `json:"archive_field"`
	ArchiveAppFilter      bool        `json:"archive_app_filter"`
	ArchiveValue          *string     `json:"archive_value"`
	UnarchiveValue        *string     `json:"unarchive_value"`
	SortField             *string     `json:"sort_field"`
	Accountability        string      `json:"accountability"`
	Color                 *string     `json:"color"`
	ItemDuplicationFields interface{} `json:"item_duplication_fields"`
	Sort                  *int        `json:"sort"`
	Group                 *string     `json:"group"`
	Collapse              string      `json:"collapse"`
	PreviewURL            *string     `json:"preview_url"`
	Versioning            bool        `json:"versioning"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
}

// Field represents a field definition in a collection
type Field struct {
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

// Request types for collections
type CreateCollectionRequest struct {
	Collection            string      `json:"collection" binding:"required"`
	Icon                  *string     `json:"icon"`
	Note                  *string     `json:"note"`
	DisplayTemplate       *string     `json:"display_template"`
	Hidden                *bool       `json:"hidden"`
	Singleton             *bool       `json:"singleton"`
	Translations          interface{} `json:"translations"`
	ArchiveField          *string     `json:"archive_field"`
	ArchiveAppFilter      *bool       `json:"archive_app_filter"`
	ArchiveValue          *string     `json:"archive_value"`
	UnarchiveValue        *string     `json:"unarchive_value"`
	SortField             *string     `json:"sort_field"`
	Accountability        *string     `json:"accountability"`
	Color                 *string     `json:"color"`
	ItemDuplicationFields interface{} `json:"item_duplication_fields"`
	Sort                  *int        `json:"sort"`
	Group                 *string     `json:"group"`
	Collapse              *string     `json:"collapse"`
	PreviewURL            *string     `json:"preview_url"`
	Versioning            *bool       `json:"versioning"`
	Fields                []Field     `json:"fields"`
}

type UpdateCollectionRequest struct {
	Icon                  *string     `json:"icon"`
	Note                  *string     `json:"note"`
	DisplayTemplate       *string     `json:"display_template"`
	Hidden                *bool       `json:"hidden"`
	Singleton             *bool       `json:"singleton"`
	Translations          interface{} `json:"translations"`
	ArchiveField          *string     `json:"archive_field"`
	ArchiveAppFilter      *bool       `json:"archive_app_filter"`
	ArchiveValue          *string     `json:"archive_value"`
	UnarchiveValue        *string     `json:"unarchive_value"`
	SortField             *string     `json:"sort_field"`
	Accountability        *string     `json:"accountability"`
	Color                 *string     `json:"color"`
	ItemDuplicationFields interface{} `json:"item_duplication_fields"`
	Sort                  *int        `json:"sort"`
	Group                 *string     `json:"group"`
	Collapse              *string     `json:"collapse"`
	PreviewURL            *string     `json:"preview_url"`
	Versioning            *bool       `json:"versioning"`
}

// isAdmin checks if the requesting user is an admin
func (h *CollectionsHandler) isAdmin(c *gin.Context) bool {
	currentUserRole := c.GetString("user_role")
	return currentUserRole == "Administrator"
}

// Collections handlers implementations
// GetCollections retrieves all collections
//
//	@Summary		Get all collections
//	@Description	Retrieve a list of all collections in the system
//	@Tags			collections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		CollectionModel	"List of collections"
//	@Failure		401	{object}	ErrorResponse	"Unauthorized"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/collections [get]
func (h *CollectionsHandler) getCollections(c *gin.Context) {
	// Parse query parameters for pagination
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

	offset := (page - 1) * limit

	// Query collections with pagination
	query := `
		SELECT collection, icon, note, display_template, hidden, singleton, 
		       translations, archive_field, archive_app_filter, archive_value, 
		       unarchive_value, sort_field, accountability, color, 
		       item_duplication_fields, sort, "group", collapse, preview_url, 
		       versioning, created_at, updated_at
		FROM collections 
		ORDER BY sort ASC, collection ASC 
		LIMIT $1 OFFSET $2
	`

	rows, err := h.db.Query(query, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Database error while fetching collections")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var collections []Collection
	for rows.Next() {
		var collection Collection
		var translationsBytes, itemDuplicationFieldsBytes []byte

		err := rows.Scan(
			&collection.Collection, &collection.Icon, &collection.Note,
			&collection.DisplayTemplate, &collection.Hidden, &collection.Singleton,
			&translationsBytes, &collection.ArchiveField, &collection.ArchiveAppFilter,
			&collection.ArchiveValue, &collection.UnarchiveValue, &collection.SortField,
			&collection.Accountability, &collection.Color, &itemDuplicationFieldsBytes,
			&collection.Sort, &collection.Group, &collection.Collapse,
			&collection.PreviewURL, &collection.Versioning, &collection.CreatedAt,
			&collection.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Error scanning collection row")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Parse JSON fields
		if translationsBytes != nil {
			json.Unmarshal(translationsBytes, &collection.Translations)
		}
		if itemDuplicationFieldsBytes != nil {
			json.Unmarshal(itemDuplicationFieldsBytes, &collection.ItemDuplicationFields)
		}

		collections = append(collections, collection)
	}

	// Get total count for pagination
	var total int
	err = h.db.QueryRow("SELECT COUNT(*) FROM collections").Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Error counting collections")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": collections,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// CreateCollection creates a new collection
//
//	@Summary		Create a new collection
//	@Description	Create a new collection in the system
//	@Tags			collections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	body		CollectionModel	true	"Collection data"
//	@Success		201			{object}	CollectionModel	"Created collection"
//	@Failure		400			{object}	ErrorResponse	"Invalid request payload"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		409			{object}	ErrorResponse	"Collection already exists"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/collections [post]
func (h *CollectionsHandler) createCollection(c *gin.Context) {
	// Only admins can create collections
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid create collection request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if collection already exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM collections WHERE collection = $1)", req.Collection).Scan(&exists)
	if err != nil {
		logrus.WithError(err).Error("Database error while checking collection existence")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Collection already exists"})
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
	hidden := false
	singleton := false
	archiveAppFilter := true
	accountability := "all"
	collapse := "open"
	versioning := false

	if req.Hidden != nil {
		hidden = *req.Hidden
	}
	if req.Singleton != nil {
		singleton = *req.Singleton
	}
	if req.ArchiveAppFilter != nil {
		archiveAppFilter = *req.ArchiveAppFilter
	}
	if req.Accountability != nil {
		accountability = *req.Accountability
	}
	if req.Collapse != nil {
		collapse = *req.Collapse
	}
	if req.Versioning != nil {
		versioning = *req.Versioning
	}

	// Convert JSON fields to bytes or NULL
	var translationsData, itemDuplicationFieldsData interface{}
	if req.Translations != nil {
		translationsBytes, _ := json.Marshal(req.Translations)
		translationsData = translationsBytes
	} else {
		translationsData = nil
	}
	if req.ItemDuplicationFields != nil {
		itemDuplicationFieldsBytes, _ := json.Marshal(req.ItemDuplicationFields)
		itemDuplicationFieldsData = itemDuplicationFieldsBytes
	} else {
		itemDuplicationFieldsData = nil
	}

	// Insert collection
	_, err = tx.Exec(`
		INSERT INTO collections (
			collection, icon, note, display_template, hidden, singleton,
			translations, archive_field, archive_app_filter, archive_value,
			unarchive_value, sort_field, accountability, color, 
			item_duplication_fields, sort, "group", collapse, preview_url, versioning
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`, req.Collection, req.Icon, req.Note, req.DisplayTemplate, hidden, singleton,
		translationsData, req.ArchiveField, archiveAppFilter, req.ArchiveValue,
		req.UnarchiveValue, req.SortField, accountability, req.Color,
		itemDuplicationFieldsData, req.Sort, req.Group, collapse, req.PreviewURL, versioning)

	if err != nil {
		logrus.WithError(err).Error("Database error while creating collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Create the actual data table for this collection
	createTableSQL := `CREATE TABLE IF NOT EXISTS "` + req.Collection + `" (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = tx.Exec(createTableSQL)
	if err != nil {
		logrus.WithError(err).Error("Failed to create collection table")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection table"})
		return
	}

	// Add fields if provided
	if len(req.Fields) > 0 {
		for _, field := range req.Fields {
			// Handle JSON fields properly - use nil for NULL instead of empty byte slices
			var optionsData, displayOptionsData, translationsData, conditionsData, validationData interface{}

			if field.Options != nil {
				optionsBytes, _ := json.Marshal(field.Options)
				optionsData = optionsBytes
			} else {
				optionsData = nil
			}
			if field.DisplayOptions != nil {
				displayOptionsBytes, _ := json.Marshal(field.DisplayOptions)
				displayOptionsData = displayOptionsBytes
			} else {
				displayOptionsData = nil
			}
			if field.Translations != nil {
				translationsBytes, _ := json.Marshal(field.Translations)
				translationsData = translationsBytes
			} else {
				translationsData = nil
			}
			if field.Conditions != nil {
				conditionsBytes, _ := json.Marshal(field.Conditions)
				conditionsData = conditionsBytes
			} else {
				conditionsData = nil
			}
			if field.Validation != nil {
				validationBytes, _ := json.Marshal(field.Validation)
				validationData = validationBytes
			} else {
				validationData = nil
			}

			// Set field defaults
			width := "full"
			if field.Width != "" {
				width = field.Width
			}

			// Convert special array to PostgreSQL array format
			var special interface{}
			if len(field.Special) > 0 {
				special = pq.Array(field.Special)
			} else {
				special = pq.Array([]string{})
			}

			_, err = tx.Exec(`
				INSERT INTO fields (
					collection, field, special, interface, options, display, 
					display_options, readonly, hidden, sort, width, translations,
					note, conditions, required, "group", validation, validation_message
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
			`, req.Collection, field.Field, special, field.Interface, optionsData,
				field.Display, displayOptionsData, field.Readonly, field.Hidden,
				field.Sort, width, translationsData, field.Note, conditionsData,
				field.Required, field.Group, validationData, field.ValidationMessage)

			if err != nil {
				logrus.WithError(err).Error("Database error while creating field")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		logrus.WithError(err).Error("Failed to commit transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Return the created collection
	collection, err := h.getCollectionByName(req.Collection)
	if err != nil {
		logrus.WithError(err).Error("Error fetching created collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithField("collection", req.Collection).Info("Collection created successfully")
	c.JSON(http.StatusCreated, gin.H{"data": collection})
}

// GetCollection retrieves a specific collection by name
//
//	@Summary		Get collection by name
//	@Description	Retrieve a specific collection by its name
//	@Tags			collections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string			true	"Collection name"
//	@Success		200			{object}	CollectionModel	"Collection details"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		404			{object}	ErrorResponse	"Collection not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/collections/{collection} [get]
func (h *CollectionsHandler) getCollection(c *gin.Context) {
	collectionName := c.Param("collection")

	collection, err := h.getCollectionByName(collectionName)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Also get fields for this collection
	fields, err := h.getFieldsByCollection(collectionName)
	if err != nil {
		logrus.WithError(err).Error("Database error while fetching collection fields")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	result := map[string]interface{}{
		"collection": collection,
		"fields":     fields,
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *CollectionsHandler) updateCollection(c *gin.Context) {
	// Only admins can update collections
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	collectionName := c.Param("collection")

	// Check if collection exists
	_, err := h.getCollectionByName(collectionName)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var req UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid update collection request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Build update query dynamically
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Icon != nil {
		updateFields = append(updateFields, "icon = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Icon)
		argIndex++
	}
	if req.Note != nil {
		updateFields = append(updateFields, "note = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Note)
		argIndex++
	}
	if req.DisplayTemplate != nil {
		updateFields = append(updateFields, "display_template = $"+strconv.Itoa(argIndex))
		args = append(args, *req.DisplayTemplate)
		argIndex++
	}
	if req.Hidden != nil {
		updateFields = append(updateFields, "hidden = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Hidden)
		argIndex++
	}
	if req.Singleton != nil {
		updateFields = append(updateFields, "singleton = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Singleton)
		argIndex++
	}
	if req.Translations != nil {
		translationsBytes, _ := json.Marshal(req.Translations)
		updateFields = append(updateFields, "translations = $"+strconv.Itoa(argIndex))
		args = append(args, translationsBytes)
		argIndex++
	}
	if req.ArchiveField != nil {
		updateFields = append(updateFields, "archive_field = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ArchiveField)
		argIndex++
	}
	if req.ArchiveAppFilter != nil {
		updateFields = append(updateFields, "archive_app_filter = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ArchiveAppFilter)
		argIndex++
	}
	if req.ArchiveValue != nil {
		updateFields = append(updateFields, "archive_value = $"+strconv.Itoa(argIndex))
		args = append(args, *req.ArchiveValue)
		argIndex++
	}
	if req.UnarchiveValue != nil {
		updateFields = append(updateFields, "unarchive_value = $"+strconv.Itoa(argIndex))
		args = append(args, *req.UnarchiveValue)
		argIndex++
	}
	if req.SortField != nil {
		updateFields = append(updateFields, "sort_field = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SortField)
		argIndex++
	}
	if req.Accountability != nil {
		updateFields = append(updateFields, "accountability = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Accountability)
		argIndex++
	}
	if req.Color != nil {
		updateFields = append(updateFields, "color = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Color)
		argIndex++
	}
	if req.ItemDuplicationFields != nil {
		itemDuplicationFieldsBytes, _ := json.Marshal(req.ItemDuplicationFields)
		updateFields = append(updateFields, "item_duplication_fields = $"+strconv.Itoa(argIndex))
		args = append(args, itemDuplicationFieldsBytes)
		argIndex++
	}
	if req.Sort != nil {
		updateFields = append(updateFields, "sort = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Sort)
		argIndex++
	}
	if req.Group != nil {
		updateFields = append(updateFields, "\"group\" = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Group)
		argIndex++
	}
	if req.Collapse != nil {
		updateFields = append(updateFields, "collapse = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Collapse)
		argIndex++
	}
	if req.PreviewURL != nil {
		updateFields = append(updateFields, "preview_url = $"+strconv.Itoa(argIndex))
		args = append(args, *req.PreviewURL)
		argIndex++
	}
	if req.Versioning != nil {
		updateFields = append(updateFields, "versioning = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Versioning)
		argIndex++
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Add updated_at and collection name to the query
	updateFields = append(updateFields, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, collectionName)

	query := "UPDATE collections SET " + strings.Join(updateFields, ", ") + " WHERE collection = $" + strconv.Itoa(argIndex)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Database error while updating collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Return updated collection
	collection, err := h.getCollectionByName(collectionName)
	if err != nil {
		logrus.WithError(err).Error("Error fetching updated collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithField("collection", collectionName).Info("Collection updated successfully")
	c.JSON(http.StatusOK, gin.H{"data": collection})
}

func (h *CollectionsHandler) deleteCollection(c *gin.Context) {
	// Only admins can delete collections
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	collectionName := c.Param("collection")

	// Check if collection exists
	_, err := h.getCollectionByName(collectionName)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Prevent deletion of system collections if any
	systemCollections := []string{"users", "roles", "permissions", "collections", "fields", "sessions", "activity", "revisions", "settings"}
	for _, sysCol := range systemCollections {
		if collectionName == sysCol {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete system collection"})
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

	// Delete related fields first
	_, err = tx.Exec("DELETE FROM fields WHERE collection = $1", collectionName)
	if err != nil {
		logrus.WithError(err).Error("Database error while deleting collection fields")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Delete collection metadata
	_, err = tx.Exec("DELETE FROM collections WHERE collection = $1", collectionName)
	if err != nil {
		logrus.WithError(err).Error("Database error while deleting collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Drop the actual data table (be careful with this)
	dropTableSQL := `DROP TABLE IF EXISTS "` + collectionName + `" CASCADE`
	_, err = tx.Exec(dropTableSQL)
	if err != nil {
		logrus.WithError(err).Error("Failed to drop collection table")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to drop collection table"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		logrus.WithError(err).Error("Failed to commit transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithField("collection", collectionName).Info("Collection deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Collection deleted successfully"})
}

// Helper method to get collection by name
func (h *CollectionsHandler) getCollectionByName(name string) (*Collection, error) {
	query := `
		SELECT collection, icon, note, display_template, hidden, singleton, 
		       translations, archive_field, archive_app_filter, archive_value, 
		       unarchive_value, sort_field, accountability, color, 
		       item_duplication_fields, sort, "group", collapse, preview_url, 
		       versioning, created_at, updated_at
		FROM collections 
		WHERE collection = $1
	`

	var collection Collection
	var translationsBytes, itemDuplicationFieldsBytes []byte

	err := h.db.QueryRow(query, name).Scan(
		&collection.Collection, &collection.Icon, &collection.Note,
		&collection.DisplayTemplate, &collection.Hidden, &collection.Singleton,
		&translationsBytes, &collection.ArchiveField, &collection.ArchiveAppFilter,
		&collection.ArchiveValue, &collection.UnarchiveValue, &collection.SortField,
		&collection.Accountability, &collection.Color, &itemDuplicationFieldsBytes,
		&collection.Sort, &collection.Group, &collection.Collapse,
		&collection.PreviewURL, &collection.Versioning, &collection.CreatedAt,
		&collection.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if translationsBytes != nil {
		json.Unmarshal(translationsBytes, &collection.Translations)
	}
	if itemDuplicationFieldsBytes != nil {
		json.Unmarshal(itemDuplicationFieldsBytes, &collection.ItemDuplicationFields)
	}

	return &collection, nil
}

// Helper method to get fields by collection
func (h *CollectionsHandler) getFieldsByCollection(collectionName string) ([]Field, error) {
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

	var fields []Field
	for rows.Next() {
		var field Field
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

		fields = append(fields, field)
	}

	return fields, nil
}
