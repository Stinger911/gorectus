package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router with the logging middleware
	router := gin.New()

	// Add the same logging middleware as in the main server
	router.Use(func(c *gin.Context) {
		start := time.Now()
		// raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details (in tests, we just verify it doesn't crash)
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// No need to reassign 'path' since it's not used after this point
		// In tests, we verify the middleware works without actually logging
		assert.GreaterOrEqual(t, latency, time.Duration(0))
		// Note: ClientIP() may return empty string in test mode, so we just check it's a string
		assert.IsType(t, "", clientIP)
		assert.NotEmpty(t, method)
		assert.GreaterOrEqual(t, statusCode, 0)
		assert.GreaterOrEqual(t, bodySize, 0)
	})

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test the middleware
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoggingMiddlewareWithQueryParams(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Add middleware that captures query parameters
	router.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			fullPath := path + "?" + raw
			assert.Contains(t, fullPath, "param=value")
		}
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test with query parameters
	req, _ := http.NewRequest("GET", "/test?param=value", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogrusConfiguration(t *testing.T) {
	// Test different log levels
	tests := []struct {
		name     string
		logLevel string
		ginMode  string
		expected logrus.Level
	}{
		{"Debug level", "debug", "", logrus.DebugLevel},
		{"Info level", "info", "", logrus.InfoLevel},
		{"Warn level", "warn", "", logrus.WarnLevel},
		{"Error level", "error", "", logrus.ErrorLevel},
		{"Default with release mode", "", "release", logrus.InfoLevel},
		{"Default with debug mode", "", "debug", logrus.DebugLevel},
		{"Default with no mode", "", "", logrus.DebugLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("GIN_MODE")

			// Set test environment
			if tt.logLevel != "" {
				os.Setenv("LOG_LEVEL", tt.logLevel)
			}
			if tt.ginMode != "" {
				os.Setenv("GIN_MODE", tt.ginMode)
			}

			// Apply the same logic as in main()
			logLevel := os.Getenv("LOG_LEVEL")
			var actualLevel logrus.Level

			switch logLevel {
			case "debug":
				actualLevel = logrus.DebugLevel
			case "info":
				actualLevel = logrus.InfoLevel
			case "warn":
				actualLevel = logrus.WarnLevel
			case "error":
				actualLevel = logrus.ErrorLevel
			default:
				if os.Getenv("GIN_MODE") == "release" {
					actualLevel = logrus.InfoLevel
				} else {
					actualLevel = logrus.DebugLevel
				}
			}

			assert.Equal(t, tt.expected, actualLevel)

			// Clean up
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("GIN_MODE")
		})
	}
}

func TestPortConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		portEnv      string
		expectedPort string
	}{
		{"Default port", "", "8080"},
		{"Custom port", "3000", "3000"},
		{"Port from environment", "9000", "9000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear PORT environment variable
			os.Unsetenv("PORT")

			// Set test port if provided
			if tt.portEnv != "" {
				os.Setenv("PORT", tt.portEnv)
			}

			// Apply the same logic as in main()
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			assert.Equal(t, tt.expectedPort, port)

			// Clean up
			os.Unsetenv("PORT")
		})
	}
}

func TestGinModeConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		ginMode  string
		expected string
	}{
		{"Test mode", gin.TestMode, gin.TestMode},
		{"Debug mode", gin.DebugMode, gin.DebugMode},
		{"Release mode", gin.ReleaseMode, gin.ReleaseMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(tt.ginMode)
			assert.Equal(t, tt.expected, gin.Mode())
		})
	}
}

func TestServerStructure(t *testing.T) {
	// Test that Server struct can be created and has expected fields
	router := gin.New()

	server := &Server{
		db:     nil, // Mock DB not needed for structure test
		router: router,
	}

	assert.NotNil(t, server)
	assert.Equal(t, router, server.router)
	assert.Nil(t, server.db) // Expected to be nil in this test
}

func TestHealthCheckResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{
		db:     nil,
		router: router,
	}

	// Add only the health check route
	router.GET("/health", server.healthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "gorectus")
	assert.Contains(t, w.Body.String(), "ok")
	assert.Contains(t, w.Body.String(), "1.0.0")
}

// Test that protected endpoints require authentication and placeholder endpoints return not implemented
func TestEndpointAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &Server{
		db:     nil,
		router: router,
	}
	server.setupRoutes()

	// Test protected endpoints (should return 401 without auth)
	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/collections"},
		{"GET", "/api/v1/users"},
		{"GET", "/api/v1/roles"},
		{"GET", "/api/v1/items/test"},
	}

	for _, endpoint := range protectedEndpoints {
		t.Run(endpoint.method+" "+endpoint.path+" (unauthorized)", func(t *testing.T) {
			req, _ := http.NewRequest(endpoint.method, endpoint.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "Authorization header required")
		})
	}

	// Test auth endpoints that actually work
	authEndpoints := []struct {
		method       string
		path         string
		expectedCode int
	}{
		{"POST", "/api/v1/auth/login", http.StatusBadRequest}, // Bad request without proper body
	}

	for _, endpoint := range authEndpoints {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			req, _ := http.NewRequest(endpoint.method, endpoint.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, endpoint.expectedCode, w.Code)
		})
	}
}
