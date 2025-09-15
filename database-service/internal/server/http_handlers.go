package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPServer wraps the database service with HTTP handlers
type HTTPServer struct {
	dbService *DatabaseService
}

// NewHTTPServer creates a new HTTP server with the database service
func NewHTTPServer(dbService *DatabaseService) *HTTPServer {
	return &HTTPServer{
		dbService: dbService,
	}
}

// SetupRoutes configures all HTTP routes
func (h *HTTPServer) SetupRoutes() *gin.Engine {
	router := gin.Default()

	router.GET("/health", h.healthCheck)

	v1 := router.Group("/api/v1")
	{
		_ = v1 // TODO: Add endpoints
	}

	return router
}

func (h *HTTPServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}