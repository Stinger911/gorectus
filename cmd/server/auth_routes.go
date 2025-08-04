package main

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related routes
type AuthHandler struct {
	db *sql.DB
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(server ServerInterface) *AuthHandler {
	return &AuthHandler{
		db: server.GetDB(),
	}
}

// SetupRoutes sets up authentication routes
func (h *AuthHandler) SetupRoutes(v1 *gin.RouterGroup) {
	// CORS preflight OPTIONS for auth endpoints
	v1.OPTIONS("/auth/login", h.optionsHandler)
	v1.OPTIONS("/auth/logout", h.optionsHandler)
	v1.OPTIONS("/auth/refresh", h.optionsHandler)

	// Authentication routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/login", h.login)
		auth.POST("/logout", h.authMiddleware(), h.logout)
		auth.POST("/refresh", h.authMiddleware(), h.refresh)
		auth.GET("/me", h.authMiddleware(), h.getCurrentUser)
	}
}

// Generic OPTIONS handler for CORS preflight requests
func (h *AuthHandler) optionsHandler(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
	c.Status(http.StatusOK)
}

// JWT Middleware for protected routes
func (h *AuthHandler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		claims, err := validateJWT(tokenString)
		if err != nil {
			logrus.WithError(err).Warn("Invalid JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func (h *AuthHandler) login(c *gin.Context) {
	var req struct {
		Username string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	c.Header("Access-Control-Allow-Origin", "*")
	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid login request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Query user from database (get ID as string, and also fetch role)
	var userID string
	var passwordHash string
	var role string
	var firstName, lastName string
	err := h.db.QueryRow(`
		SELECT u.id, u.password, r.name as role, u.first_name, u.last_name 
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE u.email = $1 AND u.status = 'active'
	`, req.Username).Scan(&userID, &passwordHash, &role, &firstName, &lastName)

	if err == sql.ErrNoRows {
		logrus.WithField("username", req.Username).Warn("Login attempt with invalid username")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error during login")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		logrus.WithField("username", req.Username).Warn("Login attempt with invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	token, err := generateJWT(userID, req.Username, role)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate JWT token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update last access time
	_, err = h.db.Exec("UPDATE users SET last_access = NOW() WHERE id = $1", userID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to update last access time")
		// Don't fail the login for this
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   req.Username,
		"role":    role,
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400, // 24 hours in seconds
		"user": gin.H{
			"id":         userID,
			"email":      req.Username,
			"first_name": firstName,
			"last_name":  lastName,
			"role":       role,
		},
	})
}

func (h *AuthHandler) logout(c *gin.Context) {
	// With JWT, logout is typically handled client-side by discarding the token
	// For server-side logout, you would need a token blacklist/revocation mechanism
	// For now, we'll just return success
	logrus.WithField("user_id", c.GetString("user_id")).Info("User logged out")
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (h *AuthHandler) refresh(c *gin.Context) {
	// Get current user from token
	userID := c.GetString("user_id")
	email := c.GetString("user_email")
	role := c.GetString("user_role")

	// Generate new token
	token, err := generateJWT(userID, email, role)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400, // 24 hours in seconds
	})
}

func (h *AuthHandler) getCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")

	var user struct {
		ID         string     `json:"id"`
		Email      string     `json:"email"`
		FirstName  string     `json:"first_name"`
		LastName   string     `json:"last_name"`
		Role       string     `json:"role"`
		Status     string     `json:"status"`
		CreatedAt  time.Time  `json:"created_at"`
		LastAccess *time.Time `json:"last_access"`
	}

	err := h.db.QueryRow(`
		SELECT u.id, u.email, u.first_name, u.last_name, r.name as role, u.status, u.created_at, u.last_access
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE u.id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Role, &user.Status, &user.CreatedAt, &user.LastAccess)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database error while fetching current user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}
