package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// InfoHandler handles edge information requests.
type InfoHandler struct{}

// NewInfoHandler creates an InfoHandler.
func NewInfoHandler() *InfoHandler {
	return &InfoHandler{}
}

// Info handles GET /api/v1/info
func (h *InfoHandler) Info(c *gin.Context) {
	clientIP := c.ClientIP()
	c.JSON(http.StatusOK, gin.H{
		"ip":        clientIP,
		"client_ip": clientIP,
		"service":   "LatinaApi",
		"version":   "1.0.0",
	})
}
