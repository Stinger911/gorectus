package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RouteHandler interface for all route handlers
type RouteHandler interface {
	SetupRoutes(group *gin.RouterGroup)
}

// Server methods needed by route handlers
type ServerInterface interface {
	GetDB() *sql.DB
	AuthMiddleware() gin.HandlerFunc
	OptionsHandler() gin.HandlerFunc
}

// Implement ServerInterface for Server
func (s *Server) GetDB() *sql.DB {
	return s.db
}

func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return s.authMiddleware()
}

func (s *Server) OptionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Status(http.StatusOK)
	}
}
