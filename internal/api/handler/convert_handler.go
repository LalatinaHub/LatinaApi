package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/LalatinaHub/LatinaApi/internal/service/converter"
	"github.com/gin-gonic/gin"
)

// ConvertHandler handles arbitrary proxy URL format conversion.
type ConvertHandler struct {
	convService converter.ConverterService
}

// NewConvertHandler creates a ConvertHandler.
func NewConvertHandler(convService converter.ConverterService) *ConvertHandler {
	return &ConvertHandler{convService: convService}
}

// Convert handles POST /api/v1/convert
func (h *ConvertHandler) Convert(c *gin.Context) {
	format := c.DefaultQuery("format", "base64")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(strings.TrimSpace(string(body))) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty proxy configuration payload"})
		return
	}

	result, err := h.convService.ConvertRaw(string(body), format)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contentType := "text/plain; charset=utf-8"
	switch strings.ToLower(format) {
	case "clash", "clash.meta", "mihomo":
		contentType = "application/yaml; charset=utf-8"
	case "singbox", "sing-box", "sfa", "bfr":
		contentType = "application/json; charset=utf-8"
	}

	c.Data(http.StatusOK, contentType, []byte(result))
}
