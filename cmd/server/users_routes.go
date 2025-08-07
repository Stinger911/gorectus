package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// UsersHandler handles user-related routes
type UsersHandler struct {
	db             *sql.DB
	authMiddleware gin.HandlerFunc
	optionsHandler gin.HandlerFunc
}

// NewUsersHandler creates a new users handler
func NewUsersHandler(server ServerInterface) *UsersHandler {
	return &UsersHandler{
		db:             server.GetDB(),
		authMiddleware: server.AuthMiddleware(),
		optionsHandler: server.OptionsHandler(),
	}
}

// SetupRoutes sets up user routes
func (h *UsersHandler) SetupRoutes(v1 *gin.RouterGroup) {
	// CORS preflight OPTIONS for users endpoints
	v1.OPTIONS("/users", h.optionsHandler)

	// Users routes (protected)
	users := v1.Group("/users")
	users.Use(h.authMiddleware)
	{
		users.GET("", h.getUsers)
		users.POST("", h.createUser)

		// It's important to register /me routes before /:id routes
		users.GET("/me", h.getMe)
		users.PATCH("/me", h.updateMe)

		users.GET("/:id", h.getUser)
		users.PATCH("/:id", h.updateUser)
		users.DELETE("/:id", h.deleteUser)
	}
}

// User represents a user in the system
type User struct {
	ID                 string     `json:"id"`
	Email              string     `json:"email"`
	FirstName          string     `json:"first_name"`
	LastName           string     `json:"last_name"`
	Avatar             *string    `json:"avatar"`
	Language           string     `json:"language"`
	Theme              string     `json:"theme"`
	Status             string     `json:"status"`
	RoleID             string     `json:"role_id"`
	RoleName           string     `json:"role_name"`
	LastAccess         *time.Time `json:"last_access"`
	LastPage           *string    `json:"last_page"`
	Provider           string     `json:"provider"`
	ExternalIdentifier *string    `json:"external_identifier"`
	EmailNotifications bool       `json:"email_notifications"`
	Tags               *string    `json:"tags"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email              string  `json:"email" binding:"required,email"`
	Password           string  `json:"password" binding:"required,min=6"`
	FirstName          string  `json:"first_name" binding:"required"`
	LastName           string  `json:"last_name" binding:"required"`
	Avatar             *string `json:"avatar"`
	Language           string  `json:"language"`
	Theme              string  `json:"theme"`
	Status             string  `json:"status"`
	RoleID             string  `json:"role_id" binding:"required"`
	EmailNotifications *bool   `json:"email_notifications"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Email              *string `json:"email" binding:"omitempty,email"`
	Password           *string `json:"password" binding:"omitempty,min=6"`
	FirstName          *string `json:"first_name"`
	LastName           *string `json:"last_name"`
	Avatar             *string `json:"avatar"`
	Language           *string `json:"language"`
	Theme              *string `json:"theme"`
	Status             *string `json:"status"`
	RoleID             *string `json:"role_id"`
	EmailNotifications *bool   `json:"email_notifications"`
}

// isAdminOrSelf checks if the requesting user is an admin or the user themselves
func (h *UsersHandler) isAdminOrSelf(c *gin.Context, targetUserID string) bool {
	currentUserID := c.GetString("user_id")
	currentUserRole := c.GetString("user_role")

	// Allow if user is admin
	if currentUserRole == "Administrator" {
		return true
	}

	// Allow if user is accessing their own data
	if currentUserID == targetUserID {
		return true
	}

	return false
}

// isAdmin checks if the requesting user is an admin
func (h *UsersHandler) isAdmin(c *gin.Context) bool {
	currentUserRole := c.GetString("user_role")
	return currentUserRole == "Administrator"
}

// Users handlers implementations

func (h *UsersHandler) getMe(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.getUserByID(userID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UsersHandler) updateMe(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid update user request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Non-admin users cannot change certain fields
	// This check is important for the /me endpoint
	if req.Status != nil || req.RoleID != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update status or role"})
		return
	}

	// Check if email already exists (if being updated)
	if req.Email != nil {
		var existingID string
		err := h.db.QueryRow("SELECT id FROM users WHERE email = $1 AND id != $2", *req.Email, userID).Scan(&existingID)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		} else if err != sql.ErrNoRows {
			logrus.WithError(err).Error("Database error while checking email")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	}

	// Build update query dynamically
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Email != nil {
		updateFields = append(updateFields, "email = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Email)
		argIndex++
	}
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.WithError(err).Error("Error hashing password")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing error"})
			return
		}
		updateFields = append(updateFields, "password = $"+strconv.Itoa(argIndex))
		args = append(args, string(hashedPassword))
		argIndex++
	}
	if req.FirstName != nil {
		updateFields = append(updateFields, "first_name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.FirstName)
		argIndex++
	}
	if req.LastName != nil {
		updateFields = append(updateFields, "last_name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.LastName)
		argIndex++
	}
	if req.Avatar != nil {
		updateFields = append(updateFields, "avatar = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Avatar)
		argIndex++
	}
	if req.Language != nil {
		updateFields = append(updateFields, "language = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Language)
		argIndex++
	}
	if req.Theme != nil {
		updateFields = append(updateFields, "theme = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Theme)
		argIndex++
	}
	if req.EmailNotifications != nil {
		updateFields = append(updateFields, "email_notifications = $"+strconv.Itoa(argIndex))
		args = append(args, *req.EmailNotifications)
		argIndex++
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Add updated_at field
	updateFields = append(updateFields, "updated_at = CURRENT_TIMESTAMP")

	// Add user ID for WHERE clause
	args = append(args, userID)

	query := "UPDATE users SET " + updateFields[0]
	for i := 1; i < len(updateFields); i++ {
		query += ", " + updateFields[i]
	}
	query += " WHERE id = $" + strconv.Itoa(argIndex)

	_, err := h.db.Exec(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Database error while updating user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Fetch updated user
	user, err := h.getUserByID(userID)
	if err != nil {
		logrus.WithError(err).Error("Error fetching updated user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"updated_by": userID,
	}).Info("User updated their own profile")

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UsersHandler) getUsers(c *gin.Context) {
	// Only admins can list all users
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

	// Query users with pagination
	query := `
		SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.language, u.theme, 
		       u.status, u.role_id, r.name as role_name, u.last_access, u.last_page, 
		       u.provider, u.external_identifier, u.email_notifications, u.tags,
		       u.created_at, u.updated_at
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		ORDER BY u.created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := h.db.Query(query, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Database error while fetching users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Avatar,
			&user.Language, &user.Theme, &user.Status, &user.RoleID, &user.RoleName,
			&user.LastAccess, &user.LastPage, &user.Provider, &user.ExternalIdentifier,
			&user.EmailNotifications, &user.Tags, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Error scanning user row")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		users = append(users, user)
	}

	// Get total count for pagination
	var total int
	err = h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Error counting users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *UsersHandler) createUser(c *gin.Context) {
	// Only admins can create users
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid create user request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if email already exists
	var existingID string
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	} else if err != sql.ErrNoRows {
		logrus.WithError(err).Error("Database error while checking email")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Error("Error hashing password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing error"})
		return
	}

	// Set defaults
	if req.Language == "" {
		req.Language = "en-US"
	}
	if req.Theme == "" {
		req.Theme = "auto"
	}
	if req.Status == "" {
		req.Status = "active"
	}
	emailNotifications := true
	if req.EmailNotifications != nil {
		emailNotifications = *req.EmailNotifications
	}

	// Insert user
	var userID string
	err = h.db.QueryRow(`
		INSERT INTO users (email, password, first_name, last_name, avatar, language, theme, 
		                  status, role_id, email_notifications)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`, req.Email, string(hashedPassword), req.FirstName, req.LastName, req.Avatar,
		req.Language, req.Theme, req.Status, req.RoleID, emailNotifications).Scan(&userID)

	if err != nil {
		logrus.WithError(err).Error("Database error while creating user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Fetch the created user
	user, err := h.getUserByID(userID)
	if err != nil {
		logrus.WithError(err).Error("Error fetching created user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"email":      req.Email,
		"created_by": c.GetString("user_id"),
	}).Info("User created successfully")

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *UsersHandler) getUser(c *gin.Context) {
	userID := c.Param("id")

	// Check authorization
	if !h.isAdminOrSelf(c, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	user, err := h.getUserByID(userID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UsersHandler) updateUser(c *gin.Context) {
	userID := c.Param("id")

	// Check authorization
	if !h.isAdminOrSelf(c, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid update user request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if user exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		logrus.WithError(err).Error("Database error while checking user existence")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Non-admin users can only update certain fields and only their own data
	currentUserRole := c.GetString("user_role")
	currentUserID := c.GetString("user_id")
	isAdmin := currentUserRole == "Administrator"
	isSelf := currentUserID == userID

	// If not admin and trying to update someone else's data, deny
	if !isAdmin && !isSelf {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Non-admin users cannot change certain fields
	if !isAdmin {
		if req.Status != nil || req.RoleID != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to update these fields"})
			return
		}
	}

	// Check if email already exists (if being updated)
	if req.Email != nil {
		var existingID string
		err := h.db.QueryRow("SELECT id FROM users WHERE email = $1 AND id != $2", *req.Email, userID).Scan(&existingID)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		} else if err != sql.ErrNoRows {
			logrus.WithError(err).Error("Database error while checking email")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	}

	// Build update query dynamically
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Email != nil {
		updateFields = append(updateFields, "email = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Email)
		argIndex++
	}
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.WithError(err).Error("Error hashing password")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing error"})
			return
		}
		updateFields = append(updateFields, "password = $"+strconv.Itoa(argIndex))
		args = append(args, string(hashedPassword))
		argIndex++
	}
	if req.FirstName != nil {
		updateFields = append(updateFields, "first_name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.FirstName)
		argIndex++
	}
	if req.LastName != nil {
		updateFields = append(updateFields, "last_name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.LastName)
		argIndex++
	}
	if req.Avatar != nil {
		updateFields = append(updateFields, "avatar = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Avatar)
		argIndex++
	}
	if req.Language != nil {
		updateFields = append(updateFields, "language = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Language)
		argIndex++
	}
	if req.Theme != nil {
		updateFields = append(updateFields, "theme = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Theme)
		argIndex++
	}
	if req.Status != nil && isAdmin {
		updateFields = append(updateFields, "status = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Status)
		argIndex++
	}
	if req.RoleID != nil && isAdmin {
		updateFields = append(updateFields, "role_id = $"+strconv.Itoa(argIndex))
		args = append(args, *req.RoleID)
		argIndex++
	}
	if req.EmailNotifications != nil {
		updateFields = append(updateFields, "email_notifications = $"+strconv.Itoa(argIndex))
		args = append(args, *req.EmailNotifications)
		argIndex++
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Add updated_at field
	updateFields = append(updateFields, "updated_at = CURRENT_TIMESTAMP")

	// Add user ID for WHERE clause
	args = append(args, userID)

	// query := "UPDATE users SET " + strconv.FormatInt(int64(len(updateFields)), 10) + " WHERE id = $" + strconv.Itoa(argIndex)
	query := "UPDATE users SET " + updateFields[0]
	for i := 1; i < len(updateFields); i++ {
		query += ", " + updateFields[i]
	}
	query += " WHERE id = $" + strconv.Itoa(argIndex)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Database error while updating user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Fetch updated user
	user, err := h.getUserByID(userID)
	if err != nil {
		logrus.WithError(err).Error("Error fetching updated user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"updated_by": c.GetString("user_id"),
	}).Info("User updated successfully")

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UsersHandler) deleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Only admins can delete users
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Prevent admin from deleting themselves
	currentUserID := c.GetString("user_id")
	if currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Check if user exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		logrus.WithError(err).Error("Database error while checking user existence")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete user
	_, err = h.db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		logrus.WithError(err).Error("Database error while deleting user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"deleted_by": currentUserID,
	}).Info("User deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Helper function to get user by ID
func (h *UsersHandler) getUserByID(userID string) (*User, error) {
	var user User
	query := `
		SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.language, u.theme, 
		       u.status, u.role_id, r.name as role_name, u.last_access, u.last_page, 
		       u.provider, u.external_identifier, u.email_notifications, u.tags,
		       u.created_at, u.updated_at
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE u.id = $1
	`
	err := h.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Avatar,
		&user.Language, &user.Theme, &user.Status, &user.RoleID, &user.RoleName,
		&user.LastAccess, &user.LastPage, &user.Provider, &user.ExternalIdentifier,
		&user.EmailNotifications, &user.Tags, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
