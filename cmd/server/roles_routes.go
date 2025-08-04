package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RolesHandler handles role-related routes
type RolesHandler struct {
	db             *sql.DB
	authMiddleware gin.HandlerFunc
	optionsHandler gin.HandlerFunc
}

// NewRolesHandler creates a new roles handler
func NewRolesHandler(server ServerInterface) *RolesHandler {
	return &RolesHandler{
		db:             server.GetDB(),
		authMiddleware: server.AuthMiddleware(),
		optionsHandler: server.OptionsHandler(),
	}
}

// SetupRoutes sets up role routes
func (h *RolesHandler) SetupRoutes(v1 *gin.RouterGroup) {
	// CORS preflight OPTIONS for roles endpoints
	v1.OPTIONS("/roles", h.optionsHandler)

	// Roles routes (protected)
	roles := v1.Group("/roles")
	roles.Use(h.authMiddleware)
	{
		roles.GET("", h.getRoles)
		roles.POST("", h.createRole)
		roles.GET("/:id", h.getRole)
		roles.PATCH("/:id", h.updateRole)
		roles.DELETE("/:id", h.deleteRole)
	}
}

// Role represents a role in the system
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Description *string   `json:"description"`
	IPAccess    []string  `json:"ip_access"`
	EnforceTFA  bool      `json:"enforce_tfa"`
	AdminAccess bool      `json:"admin_access"`
	AppAccess   bool      `json:"app_access"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateRoleRequest represents the request body for creating a role
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Icon        string   `json:"icon"`
	Description *string  `json:"description"`
	IPAccess    []string `json:"ip_access"`
	EnforceTFA  *bool    `json:"enforce_tfa"`
	AdminAccess *bool    `json:"admin_access"`
	AppAccess   *bool    `json:"app_access"`
}

// UpdateRoleRequest represents the request body for updating a role
type UpdateRoleRequest struct {
	Name        *string  `json:"name"`
	Icon        *string  `json:"icon"`
	Description *string  `json:"description"`
	IPAccess    []string `json:"ip_access"`
	EnforceTFA  *bool    `json:"enforce_tfa"`
	AdminAccess *bool    `json:"admin_access"`
	AppAccess   *bool    `json:"app_access"`
}

// isAdmin checks if the requesting user is an admin
func (h *RolesHandler) isAdmin(c *gin.Context) bool {
	currentUserRole := c.GetString("user_role")
	return currentUserRole == "Administrator"
}

// Roles handlers implementations
func (h *RolesHandler) getRoles(c *gin.Context) {
	// Only admins can list all roles
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
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

	// Query roles with pagination
	query := `
		SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,
		       created_at, updated_at
		FROM roles 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := h.db.Query(query, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Database error while fetching roles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var ipAccessArray []string
		err := rows.Scan(
			&role.ID, &role.Name, &role.Icon, &role.Description, &ipAccessArray,
			&role.EnforceTFA, &role.AdminAccess, &role.AppAccess,
			&role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Error scanning role row")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		role.IPAccess = ipAccessArray
		roles = append(roles, role)
	}

	// Get total count for pagination
	var total int
	err = h.db.QueryRow("SELECT COUNT(*) FROM roles").Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Error counting roles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": roles,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *RolesHandler) createRole(c *gin.Context) {
	// Only admins can create roles
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid create role request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if role name already exists
	var existingID string
	err := h.db.QueryRow("SELECT id FROM roles WHERE name = $1", req.Name).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Role name already exists"})
		return
	} else if err != sql.ErrNoRows {
		logrus.WithError(err).Error("Database error while checking role name")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Set defaults
	if req.Icon == "" {
		req.Icon = "supervised_user_circle"
	}
	enforceTFA := false
	if req.EnforceTFA != nil {
		enforceTFA = *req.EnforceTFA
	}
	adminAccess := false
	if req.AdminAccess != nil {
		adminAccess = *req.AdminAccess
	}
	appAccess := true
	if req.AppAccess != nil {
		appAccess = *req.AppAccess
	}

	// Ensure IPAccess is not nil
	if req.IPAccess == nil {
		req.IPAccess = []string{}
	}

	// Insert role
	var roleID string
	err = h.db.QueryRow(`
		INSERT INTO roles (name, icon, description, ip_access, enforce_tfa, admin_access, app_access)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, req.Name, req.Icon, req.Description, req.IPAccess, enforceTFA, adminAccess, appAccess).Scan(&roleID)

	if err != nil {
		logrus.WithError(err).Error("Database error while creating role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Fetch the created role
	role, err := h.getRoleByID(roleID)
	if err != nil {
		logrus.WithError(err).Error("Error fetching created role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"role_id":    roleID,
		"role_name":  req.Name,
		"created_by": c.GetString("user_id"),
	}).Info("Role created successfully")

	c.JSON(http.StatusCreated, gin.H{"data": role})
}

func (h *RolesHandler) getRole(c *gin.Context) {
	// Only admins can view roles
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	roleID := c.Param("id")

	role, err := h.getRoleByID(roleID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": role})
}

// getRoleByID is a helper method to fetch a role by ID
func (h *RolesHandler) getRoleByID(roleID string) (*Role, error) {
	query := `
		SELECT id, name, icon, description, ip_access, enforce_tfa, admin_access, app_access,
		       created_at, updated_at
		FROM roles 
		WHERE id = $1
	`

	var role Role
	var ipAccessArray []string
	err := h.db.QueryRow(query, roleID).Scan(
		&role.ID, &role.Name, &role.Icon, &role.Description, &ipAccessArray,
		&role.EnforceTFA, &role.AdminAccess, &role.AppAccess,
		&role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	role.IPAccess = ipAccessArray
	return &role, nil
}

func (h *RolesHandler) updateRole(c *gin.Context) {
	// Only admins can update roles
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	roleID := c.Param("id")

	// Check if role exists
	existingRole, err := h.getRoleByID(roleID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid update role request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Prevent updating the Administrator role name to avoid breaking auth
	if existingRole.Name == "Administrator" && req.Name != nil && *req.Name != "Administrator" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot change Administrator role name"})
		return
	}

	// Check if new name already exists (if name is being updated)
	if req.Name != nil && *req.Name != existingRole.Name {
		var existingID string
		err := h.db.QueryRow("SELECT id FROM roles WHERE name = $1 AND id != $2", *req.Name, roleID).Scan(&existingID)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Role name already exists"})
			return
		} else if err != sql.ErrNoRows {
			logrus.WithError(err).Error("Database error while checking role name")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	}

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		updateFields = append(updateFields, "name = $"+strconv.Itoa(argCount))
		args = append(args, *req.Name)
		argCount++
	}
	if req.Icon != nil {
		updateFields = append(updateFields, "icon = $"+strconv.Itoa(argCount))
		args = append(args, *req.Icon)
		argCount++
	}
	if req.Description != nil {
		updateFields = append(updateFields, "description = $"+strconv.Itoa(argCount))
		args = append(args, *req.Description)
		argCount++
	}
	if req.IPAccess != nil {
		updateFields = append(updateFields, "ip_access = $"+strconv.Itoa(argCount))
		args = append(args, req.IPAccess)
		argCount++
	}
	if req.EnforceTFA != nil {
		updateFields = append(updateFields, "enforce_tfa = $"+strconv.Itoa(argCount))
		args = append(args, *req.EnforceTFA)
		argCount++
	}
	if req.AdminAccess != nil {
		updateFields = append(updateFields, "admin_access = $"+strconv.Itoa(argCount))
		args = append(args, *req.AdminAccess)
		argCount++
	}
	if req.AppAccess != nil {
		updateFields = append(updateFields, "app_access = $"+strconv.Itoa(argCount))
		args = append(args, *req.AppAccess)
		argCount++
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Add updated_at field
	updateFields = append(updateFields, "updated_at = CURRENT_TIMESTAMP")

	// Add role ID as the last parameter
	args = append(args, roleID)

	query := "UPDATE roles SET " +
		updateFields[0]
	for i := 1; i < len(updateFields); i++ {
		query += ", " + updateFields[i]
	}
	query += " WHERE id = $" + strconv.Itoa(argCount)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Database error while updating role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Fetch the updated role
	role, err := h.getRoleByID(roleID)
	if err != nil {
		logrus.WithError(err).Error("Error fetching updated role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"role_id":    roleID,
		"updated_by": c.GetString("user_id"),
	}).Info("Role updated successfully")

	c.JSON(http.StatusOK, gin.H{"data": role})
}

func (h *RolesHandler) deleteRole(c *gin.Context) {
	// Only admins can delete roles
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	roleID := c.Param("id")

	// Check if role exists
	existingRole, err := h.getRoleByID(roleID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Prevent deletion of essential system roles
	if existingRole.Name == "Administrator" || existingRole.Name == "Public" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete system roles"})
		return
	}

	// Check if any users are assigned to this role
	var userCount int
	err = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE role_id = $1", roleID).Scan(&userCount)
	if err != nil {
		logrus.WithError(err).Error("Database error while checking role usage")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if userCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":       "Cannot delete role that is assigned to users",
			"users_count": userCount,
		})
		return
	}

	// Delete the role (permissions will be deleted automatically due to CASCADE)
	_, err = h.db.Exec("DELETE FROM roles WHERE id = $1", roleID)
	if err != nil {
		logrus.WithError(err).Error("Database error while deleting role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"role_id":    roleID,
		"role_name":  existingRole.Name,
		"deleted_by": c.GetString("user_id"),
	}).Info("Role deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}
