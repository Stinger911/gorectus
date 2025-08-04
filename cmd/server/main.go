package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Server struct {
	db     *sql.DB
	router *gin.Engine
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

		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", s.login)
			auth.POST("/logout", s.logout)
			auth.POST("/refresh", s.refresh)
		}

		// Collections routes (protected)
		collections := v1.Group("/collections")
		{
			collections.GET("", s.getCollections)
			collections.POST("", s.createCollection)
			collections.GET("/:collection", s.getCollection)
			collections.PATCH("/:collection", s.updateCollection)
			collections.DELETE("/:collection", s.deleteCollection)
		}

		// Items routes (protected)
		items := v1.Group("/items")
		{
			items.GET("/:collection", s.getItems)
			items.POST("/:collection", s.createItem)
			items.GET("/:collection/:id", s.getItem)
			items.PATCH("/:collection/:id", s.updateItem)
			items.DELETE("/:collection/:id", s.deleteItem)
		}

		// Users routes (protected)
		users := v1.Group("/users")
		{
			users.GET("", s.getUsers)
			users.POST("", s.createUser)
			users.GET("/:id", s.getUser)
			users.PATCH("/:id", s.updateUser)
			users.DELETE("/:id", s.deleteUser)
		}

		// Roles routes (protected)
		roles := v1.Group("/roles")
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

// Authentication handlers (placeholder implementations)
func (s *Server) login(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Login endpoint not implemented yet"})
}

func (s *Server) logout(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Logout endpoint not implemented yet"})
}

func (s *Server) refresh(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Refresh endpoint not implemented yet"})
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
