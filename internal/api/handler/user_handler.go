package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/LatinaApi/internal/service/user"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user management requests.
type UserHandler struct {
	userService user.UserService
}

// NewUserHandler returns a new UserHandler instance.
func NewUserHandler(userService user.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetUser handles GET /user/:apiToken/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	apiToken := c.Param("apiToken")
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	u, err := h.userService.GetUser(c.Request.Context(), apiToken, id)
	if err != nil {
		if errors.Is(err, model.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized API token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, u)
}
