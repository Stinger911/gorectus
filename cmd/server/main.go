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

		// Initialize route handlers
		authHandler := NewAuthHandler(s)
		collectionsHandler := NewCollectionsHandler(s)
		itemsHandler := NewItemsHandler(s)
		usersHandler := NewUsersHandler(s)
		rolesHandler := NewRolesHandler(s)

		// Setup routes for each handler
		authHandler.SetupRoutes(v1)
		collectionsHandler.SetupRoutes(v1)
		itemsHandler.SetupRoutes(v1)
		usersHandler.SetupRoutes(v1)
		rolesHandler.SetupRoutes(v1)
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
