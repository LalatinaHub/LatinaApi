package middleware

import (
	"errors"
	"net/http"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/gin-gonic/gin"
)

// ErrorMiddleware catches pending context errors and formats standard JSON responses.
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			switch {
			case errors.Is(err, model.ErrUserNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			case errors.Is(err, model.ErrSubscriptionExpired):
				c.JSON(http.StatusForbidden, gin.H{"error": "subscription expired"})
			case errors.Is(err, model.ErrQuotaExceeded):
				c.JSON(http.StatusForbidden, gin.H{"error": "subscription quota exceeded"})
			case errors.Is(err, model.ErrInvalidToken):
				c.JSON(http.StatusForbidden, gin.H{"error": "invalid or missing token"})
			case errors.Is(err, model.ErrUnauthorized):
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case errors.Is(err, model.ErrServerNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}
	}
}
