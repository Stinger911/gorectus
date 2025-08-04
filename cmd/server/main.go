package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	db     *sql.DB
	router *gin.Engine
}

// JWT Claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWT helper functions
func generateJWT(userID, email, role string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // fallback for development
	}

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gorectus",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func validateJWT(tokenString string) (*JWTClaims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // fallback for development
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// JWT Middleware for protected routes
func (s *Server) authMiddleware() gin.HandlerFunc {
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

func main() {
	// Configure logrus
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// Set log level based on environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		if os.Getenv("GIN_MODE") == "release" {
			logrus.SetLevel(logrus.InfoLevel)
		} else {
			logrus.SetLevel(logrus.DebugLevel)
		}
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using system environment variables")
	}

	// Initialize server
	server, err := NewServer()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create server")
	}
	defer server.db.Close()

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	logrus.WithField("port", port).Info("Starting gorectus server")
	if err := server.router.Run(":" + port); err != nil {
		logrus.WithError(err).Fatal("Failed to start server")
	}
}

func NewServer() (*Server, error) {
	// Initialize database connection
	db, err := initDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Add logrus middleware for HTTP request logging
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		entry := logrus.WithFields(logrus.Fields{
			"status":    statusCode,
			"latency":   latency,
			"client_ip": clientIP,
			"method":    method,
			"path":      path,
			"body_size": bodySize,
		})

		if statusCode >= 500 {
			entry.Error("HTTP request completed with server error")
		} else if statusCode >= 400 {
			entry.Warn("HTTP request completed with client error")
		} else {
			entry.Info("HTTP request completed")
		}
	})

	// Create server instance
	server := &Server{
		db:     db,
		router: router,
	}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

func initDB() (*sql.DB, error) {
	// Build connection string from environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "gorectus"
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"host":   host,
		"port":   port,
		"dbname": dbname,
	}).Info("Successfully connected to database")
	return db, nil
}

func (s *Server) setupRoutes() {
	// API version 1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Health check endpoint
		v1.GET("/health", s.healthCheck)

		// CORS preflight OPTIONS for all POST endpoints
		v1.OPTIONS("/auth/login", s.optionsHandler)
		v1.OPTIONS("/auth/logout", s.optionsHandler)
		v1.OPTIONS("/auth/refresh", s.optionsHandler)
		v1.OPTIONS("/collections", s.optionsHandler)
		v1.OPTIONS("/items/:collection", s.optionsHandler)
		v1.OPTIONS("/users", s.optionsHandler)
		v1.OPTIONS("/roles", s.optionsHandler)

		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", s.login)
			auth.POST("/logout", s.authMiddleware(), s.logout)
			auth.POST("/refresh", s.authMiddleware(), s.refresh)
			auth.GET("/me", s.authMiddleware(), s.getCurrentUser)
		}

		// Collections routes (protected)
		collections := v1.Group("/collections")
		collections.Use(s.authMiddleware())
		{
			collections.GET("", s.getCollections)
			collections.POST("", s.createCollection)
			collections.GET("/:collection", s.getCollection)
			collections.PATCH("/:collection", s.updateCollection)
			collections.DELETE("/:collection", s.deleteCollection)
		}

		// Items routes (protected)
		items := v1.Group("/items")
		items.Use(s.authMiddleware())
		{
			items.GET("/:collection", s.getItems)
			items.POST("/:collection", s.createItem)
			items.GET("/:collection/:id", s.getItem)
			items.PATCH("/:collection/:id", s.updateItem)
			items.DELETE("/:collection/:id", s.deleteItem)
		}

		// Users routes (protected)
		users := v1.Group("/users")
		users.Use(s.authMiddleware())
		{
			users.GET("", s.getUsers)
			users.POST("", s.createUser)
			users.GET("/:id", s.getUser)
			users.PATCH("/:id", s.updateUser)
			users.DELETE("/:id", s.deleteUser)
		}

		// Roles routes (protected)
		roles := v1.Group("/roles")
		roles.Use(s.authMiddleware())
		{
			roles.GET("", s.getRoles)
			roles.POST("", s.createRole)
			roles.GET("/:id", s.getRole)
			roles.PATCH("/:id", s.updateRole)
			roles.DELETE("/:id", s.deleteRole)
		}
	}

	// Static files and admin interface
	s.router.Static("/admin", "./frontend/build")
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin")
	})
}

// Health check endpoint
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "gorectus",
		"version": "1.0.0",
	})
}

// Authentication handlers
func (s *Server) login(c *gin.Context) {
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
	err := s.db.QueryRow(`
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
	_, err = s.db.Exec("UPDATE users SET last_access = NOW() WHERE id = $1", userID)
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

func (s *Server) logout(c *gin.Context) {
	// With JWT, logout is typically handled client-side by discarding the token
	// For server-side logout, you would need a token blacklist/revocation mechanism
	// For now, we'll just return success
	logrus.WithField("user_id", c.GetString("user_id")).Info("User logged out")
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (s *Server) refresh(c *gin.Context) {
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

func (s *Server) getCurrentUser(c *gin.Context) {
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

	err := s.db.QueryRow(`
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

// Collections handlers (placeholder implementations)
func (s *Server) getCollections(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get collections endpoint not implemented yet"})
}

func (s *Server) createCollection(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Create collection endpoint not implemented yet"})
}

func (s *Server) getCollection(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get collection endpoint not implemented yet"})
}

func (s *Server) updateCollection(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update collection endpoint not implemented yet"})
}

func (s *Server) deleteCollection(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete collection endpoint not implemented yet"})
}

// Items handlers (placeholder implementations)
func (s *Server) getItems(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get items endpoint not implemented yet"})
}

func (s *Server) createItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Create item endpoint not implemented yet"})
}

func (s *Server) getItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get item endpoint not implemented yet"})
}

func (s *Server) updateItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update item endpoint not implemented yet"})
}

func (s *Server) deleteItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete item endpoint not implemented yet"})
}

// Users handlers (placeholder implementations)
func (s *Server) getUsers(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get users endpoint not implemented yet"})
}

func (s *Server) createUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Create user endpoint not implemented yet"})
}

func (s *Server) getUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get user endpoint not implemented yet"})
}

func (s *Server) updateUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update user endpoint not implemented yet"})
}

func (s *Server) deleteUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete user endpoint not implemented yet"})
}

// Roles handlers (placeholder implementations)
func (s *Server) getRoles(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get roles endpoint not implemented yet"})
}

// Generic OPTIONS handler for CORS preflight requests
func (s *Server) optionsHandler(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
	c.Status(http.StatusOK)
}

func (s *Server) createRole(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Create role endpoint not implemented yet"})
}

func (s *Server) getRole(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get role endpoint not implemented yet"})
}

func (s *Server) updateRole(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update role endpoint not implemented yet"})
}

func (s *Server) deleteRole(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete role endpoint not implemented yet"})
}
