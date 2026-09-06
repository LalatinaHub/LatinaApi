package handler

import (
	"net/http"
	"time"

	"github.com/LalatinaHub/common/database"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check and ping endpoints.
type HealthHandler struct{}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Ping handles /ping and /api/v1/ping
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// Health handles /health
func (h *HealthHandler) Health(c *gin.Context) {
	dbStatus := "healthy"
	if err := database.Ping(c.Request.Context()); err != nil {
		dbStatus = "unreachable"
	}

	statusCode := http.StatusOK
	if dbStatus != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":   "ok",
		"database": dbStatus,
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}
