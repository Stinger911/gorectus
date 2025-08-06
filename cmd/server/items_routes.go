package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ItemsHandler handles item-related routes
type ItemsHandler struct {
	db             *sql.DB
	authMiddleware gin.HandlerFunc
	optionsHandler gin.HandlerFunc
}

// NewItemsHandler creates a new items handler
func NewItemsHandler(server ServerInterface) *ItemsHandler {
	return &ItemsHandler{
		db:             server.GetDB(),
		authMiddleware: server.AuthMiddleware(),
		optionsHandler: server.OptionsHandler(),
	}
}

// SetupRoutes sets up item routes
func (h *ItemsHandler) SetupRoutes(v1 *gin.RouterGroup) {
	// CORS preflight OPTIONS for items endpoints
	v1.OPTIONS("/items/:collection", h.optionsHandler)

	// Items routes (protected)
	items := v1.Group("/items")
	items.Use(h.authMiddleware)
	{
		items.GET("/:collection", h.getItems)
		items.POST("/:collection", h.createItem)
		items.GET("/:collection/:id", h.getItem)
		items.PATCH("/:collection/:id", h.updateItem)
		items.DELETE("/:collection/:id", h.deleteItem)
	}
}

// Item represents a generic item in any collection
type Item map[string]interface{}

// FieldInfo represents basic field information needed for validation
type FieldInfo struct {
	Field    string `json:"field"`
	Required bool   `json:"required"`
}

// isAdmin checks if the requesting user is an admin
func (h *ItemsHandler) isAdmin(c *gin.Context) bool {
	currentUserRole := c.GetString("user_role")
	return currentUserRole == "Administrator"
}

// checkCollectionExists verifies if a collection exists
func (h *ItemsHandler) checkCollectionExists(collectionName string) error {
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM collections WHERE collection = $1)", collectionName).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}
	return nil
}

// Items handlers implementations
// GetItems retrieves all items from a collection
//
//	@Summary		Get all items from a collection
//	@Description	Retrieve a list of all items from a specific collection
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string		true	"Collection name"
//	@Param			limit		query		int			false	"Limit the number of results"
//	@Param			offset		query		int			false	"Offset for pagination"
//	@Success		200			{array}		ItemModel	"List of items"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		404			{object}	ErrorResponse	"Collection not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/items/{collection} [get]
func (h *ItemsHandler) getItems(c *gin.Context) {
	collectionName := c.Param("collection")

	// Check if collection exists
	if err := h.checkCollectionExists(collectionName); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while checking collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

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

	// Build query - use safe table name quoting
	query := fmt.Sprintf(`SELECT * FROM "%s" ORDER BY created_at DESC LIMIT $1 OFFSET $2`, collectionName)

	rows, err := h.db.Query(query, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Database error while fetching items")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		logrus.WithError(err).Error("Error getting column names")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var items []Item
	for rows.Next() {
		// Create a slice of interface{} to receive the row data
		values := make([]interface{}, len(columns))
		valuePointers := make([]interface{}, len(columns))
		for i := range columns {
			valuePointers[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePointers...); err != nil {
			logrus.WithError(err).Error("Error scanning item row")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Create the item map
		item := make(Item)
		for i, col := range columns {
			val := values[i]
			if val != nil {
				// Handle different PostgreSQL types
				switch v := val.(type) {
				case []byte:
					// Try to parse as JSON first, fallback to string
					var jsonVal interface{}
					if err := json.Unmarshal(v, &jsonVal); err == nil {
						item[col] = jsonVal
					} else {
						item[col] = string(v)
					}
				case time.Time:
					item[col] = v.Format(time.RFC3339)
				default:
					item[col] = v
				}
			} else {
				item[col] = nil
			}
		}

		items = append(items, item)
	}

	// Get total count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, collectionName)
	var total int
	err = h.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Error counting items")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// CreateItem creates a new item in a collection
//
//	@Summary		Create a new item
//	@Description	Create a new item in a specific collection
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string		true	"Collection name"
//	@Param			item		body		ItemModel	true	"Item data"
//	@Success		201			{object}	ItemModel	"Created item"
//	@Failure		400			{object}	ErrorResponse	"Invalid request payload"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		404			{object}	ErrorResponse	"Collection not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/items/{collection} [post]
func (h *ItemsHandler) createItem(c *gin.Context) {
	// Only admins can create items
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	collectionName := c.Param("collection")

	// Check if collection exists
	if err := h.checkCollectionExists(collectionName); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while checking collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var requestData Item
	if err := c.ShouldBindJSON(&requestData); err != nil {
		logrus.WithError(err).Error("Invalid create item request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Remove system fields that shouldn't be set by user
	delete(requestData, "id")
	delete(requestData, "created_at")
	delete(requestData, "updated_at")

	if len(requestData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No data provided"})
		return
	}

	// Get fields for this collection to validate the data
	fields, err := h.getFieldsByCollection(collectionName)
	if err != nil {
		logrus.WithError(err).Error("Error getting collection fields")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Validate required fields
	for _, field := range fields {
		if field.Required && requestData[field.Field] == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Required field '%s' is missing", field.Field)})
			return
		}
	}

	// Build insert query
	columns := make([]string, 0, len(requestData))
	placeholders := make([]string, 0, len(requestData))
	values := make([]interface{}, 0, len(requestData))
	argIndex := 1

	for col, val := range requestData {
		columns = append(columns, fmt.Sprintf(`"%s"`, col))
		placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))

		// Convert complex types to JSON
		if val != nil {
			switch v := val.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, _ := json.Marshal(v)
				values = append(values, jsonBytes)
			default:
				values = append(values, v)
			}
		} else {
			values = append(values, nil)
		}

		argIndex++
	}

	insertQuery := fmt.Sprintf(
		`INSERT INTO "%s" (%s) VALUES (%s) RETURNING id, created_at, updated_at`,
		collectionName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	var newID string
	var createdAt, updatedAt time.Time
	err = h.db.QueryRow(insertQuery, values...).Scan(&newID, &createdAt, &updatedAt)
	if err != nil {
		logrus.WithError(err).Error("Database error while creating item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Get the created item
	item, err := h.getItemByID(collectionName, newID)
	if err != nil {
		logrus.WithError(err).Error("Error fetching created item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"collection": collectionName,
		"item_id":    newID,
	}).Info("Item created successfully")

	c.JSON(http.StatusCreated, gin.H{"data": item})
}

// GetItem retrieves a specific item from a collection
//
//	@Summary		Get item by ID
//	@Description	Retrieve a specific item from a collection by its ID
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string		true	"Collection name"
//	@Param			id			path		string		true	"Item ID"
//	@Success		200			{object}	ItemModel	"Item details"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		404			{object}	ErrorResponse	"Item not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/items/{collection}/{id} [get]
func (h *ItemsHandler) getItem(c *gin.Context) {
	collectionName := c.Param("collection")
	itemID := c.Param("id")

	// Check if collection exists
	if err := h.checkCollectionExists(collectionName); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while checking collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	item, err := h.getItemByID(collectionName, itemID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

// updateItem updates an existing item in a collection
//
//	@Summary		Update an existing item
//	@Description	Update an existing item in a specific collection
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string		true	"Collection name"
//	@Param			id			path		string		true	"Item ID"
//	@Param			item		body		ItemModel	true	"Updated item data"
//	@Success		200			{object}	ItemModel	"Updated item"
//	@Failure		400			{object}	ErrorResponse	"Invalid request payload"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		404			{object}	ErrorResponse	"Item not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/items/{collection}/{id} [patch]
func (h *ItemsHandler) updateItem(c *gin.Context) {
	// Only admins can update items
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}
	collectionName := c.Param("collection")
	itemID := c.Param("id")

	// Check if collection exists
	if err := h.checkCollectionExists(collectionName); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while checking collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if item exists
	_, err := h.getItemByID(collectionName, itemID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while checking item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var requestData Item
	if err := c.ShouldBindJSON(&requestData); err != nil {
		logrus.WithError(err).Error("Invalid update item request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Remove system fields that shouldn't be updated by user
	delete(requestData, "id")
	delete(requestData, "created_at")
	delete(requestData, "updated_at")

	if len(requestData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No data provided for update"})
		return
	}

	// Build update query
	updateFields := make([]string, 0, len(requestData))
	values := make([]interface{}, 0, len(requestData)+1)
	argIndex := 1

	for col, val := range requestData {
		updateFields = append(updateFields, fmt.Sprintf(`"%s" = $%d`, col, argIndex))

		// Convert complex types to JSON
		if val != nil {
			switch v := val.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, _ := json.Marshal(v)
				values = append(values, jsonBytes)
			default:
				values = append(values, v)
			}
		} else {
			values = append(values, nil)
		}

		argIndex++
	}

	// Add updated_at
	updateFields = append(updateFields, "updated_at = CURRENT_TIMESTAMP")
	values = append(values, itemID)

	updateQuery := fmt.Sprintf(
		`UPDATE "%s" SET %s WHERE id = $%d`,
		collectionName,
		strings.Join(updateFields, ", "),
		argIndex,
	)

	_, err = h.db.Exec(updateQuery, values...)
	if err != nil {
		logrus.WithError(err).Error("Database error while updating item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Get updated item
	item, err := h.getItemByID(collectionName, itemID)
	if err != nil {
		logrus.WithError(err).Error("Error fetching updated item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"collection": collectionName,
		"item_id":    itemID,
	}).Info("Item updated successfully")

	c.JSON(http.StatusOK, gin.H{"data": item})
}

// deleteItem deletes an item from a collection
//
//	@Summary		Delete an item
//	@Description	Delete a specific item from a collection by its ID
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			collection	path		string		true	"Collection name"
//	@Param			id			path		string		true	"Item ID"
//	@Success		200			{object}	SuccessMessage	"Success message"
//	@Failure		401			{object}	ErrorResponse	"Unauthorized"
//	@Failure		404			{object}	ErrorResponse	"Item not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/items/{collection}/{id} [delete]
func (h *ItemsHandler) deleteItem(c *gin.Context) {
	// Only admins can delete items
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	collectionName := c.Param("collection")
	itemID := c.Param("id")

	// Check if collection exists
	if err := h.checkCollectionExists(collectionName); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while checking collection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if item exists
	_, err := h.getItemByID(collectionName, itemID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while checking item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Delete item
	deleteQuery := fmt.Sprintf(`DELETE FROM "%s" WHERE id = $1`, collectionName)
	_, err = h.db.Exec(deleteQuery, itemID)
	if err != nil {
		logrus.WithError(err).Error("Database error while deleting item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"collection": collectionName,
		"item_id":    itemID,
	}).Info("Item deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}

// Helper method to get fields by collection
func (h *ItemsHandler) getFieldsByCollection(collectionName string) ([]FieldInfo, error) {
	query := `
		SELECT field, required
		FROM fields 
		WHERE collection = $1
		ORDER BY sort ASC, field ASC
	`

	rows, err := h.db.Query(query, collectionName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []FieldInfo
	for rows.Next() {
		var field FieldInfo
		err := rows.Scan(&field.Field, &field.Required)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	return fields, nil
}

// Helper method to get item by ID
func (h *ItemsHandler) getItemByID(collectionName, itemID string) (Item, error) {
	query := fmt.Sprintf(`SELECT * FROM "%s" WHERE id = $1`, collectionName)

	rows, err := h.db.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Create a slice of interface{} to receive the row data
	values := make([]interface{}, len(columns))
	valuePointers := make([]interface{}, len(columns))
	for i := range columns {
		valuePointers[i] = &values[i]
	}

	// Scan the row into the value pointers
	if err := rows.Scan(valuePointers...); err != nil {
		return nil, err
	}

	// Create the item map
	item := make(Item)
	for i, col := range columns {
		val := values[i]
		if val != nil {
			// Handle different PostgreSQL types
			switch v := val.(type) {
			case []byte:
				// Try to parse as JSON first, fallback to string
				var jsonVal interface{}
				if err := json.Unmarshal(v, &jsonVal); err == nil {
					item[col] = jsonVal
				} else {
					item[col] = string(v)
				}
			case time.Time:
				item[col] = v.Format(time.RFC3339)
			default:
				item[col] = v
			}
		} else {
			item[col] = nil
		}
	}

	return item, nil
}
