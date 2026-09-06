package handler

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/LatinaApi/internal/service/dbadmin"
	"github.com/gin-gonic/gin"
)

// AdminHandler handles administrative database executions.
type AdminHandler struct {
	adminService dbadmin.DBAdminService
}

// NewAdminHandler returns a new AdminHandler instance.
func NewAdminHandler(adminService dbadmin.DBAdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

type execPayload struct {
	Queries []string `json:"queries"`
}

// ExecSQL handles POST /db/:apiToken/exec
func (h *AdminHandler) ExecSQL(c *gin.Context) {
	apiToken := c.Param("apiToken")
	if apiToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "apiToken required"})
		return
	}

	var queries []string

	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var payload execPayload
		if err := c.ShouldBindJSON(&payload); err == nil && len(payload.Queries) > 0 {
			queries = payload.Queries
		}
	}

	// Fallback to reading raw body line by line
	if len(queries) == 0 {
		body, err := io.ReadAll(c.Request.Body)
		if err == nil && len(body) > 0 {
			scanner := bufio.NewScanner(strings.NewReader(string(body)))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" && !strings.HasPrefix(line, "--") {
					queries = append(queries, line)
				}
			}
		}
	}

	if len(queries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no SQL queries provided in request body"})
		return
	}

	results, err := h.adminService.ExecSQL(c.Request.Context(), apiToken, queries)
	if err != nil {
		if errors.Is(err, model.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized API token"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"executed": len(queries),
		"results":  results,
	})
}
