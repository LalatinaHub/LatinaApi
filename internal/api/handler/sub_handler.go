package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/LatinaApi/internal/service/subscription"
	"github.com/gin-gonic/gin"
)

// SubHandler handles subscription requests for clients.
type SubHandler struct {
	subService subscription.SubscriptionService
}

// NewSubHandler returns a new SubHandler instance.
func NewSubHandler(subService subscription.SubscriptionService) *SubHandler {
	return &SubHandler{subService: subService}
}

// HandleSub processes GET /sub and GET /api/v1/sub requests.
func (h *SubHandler) HandleSub(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		token = c.Query("pass") // backward-compatible alias
	}

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameter token is required"})
		return
	}

	var tlsPtr *bool
	if tlsStr := c.Query("tls"); tlsStr != "" {
		val := tlsStr == "1" || tlsStr == "true"
		tlsPtr = &val
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	free := c.Query("free") == "1" || c.Query("free") == "true"
	premium := c.Query("premium") == "1" || c.Query("premium") == "true"

	req := subscription.SubscriptionRequest{
		Token:       token,
		VPN:         c.Query("vpn"),
		Format:      c.Query("format"),
		UserAgent:   c.GetHeader("User-Agent"),
		Region:      c.Query("region"),
		CountryCode: c.Query("cc"),
		Mode:        c.Query("mode"),
		Transport:   c.Query("transport"),
		TLS:         tlsPtr,
		Free:        free,
		Premium:     premium,
		Limit:       limit,
		CDN:         c.Query("cdn"),
		SNI:         c.Query("sni"),
		Include:     c.Query("include"),
		Exclude:     c.Query("exclude"),
	}

	res, err := h.subService.GetSubscription(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, model.ErrSubscriptionExpired):
			c.JSON(http.StatusForbidden, gin.H{"error": "subscription expired"})
		case errors.Is(err, model.ErrQuotaExceeded):
			c.JSON(http.StatusForbidden, gin.H{"error": "subscription quota exceeded"})
		case errors.Is(err, model.ErrInvalidToken):
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid token"})
		case errors.Is(err, model.ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"error": "subscription inactive or unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Set standard international subscription headers
	c.Header("Subscription-Userinfo", res.UserInfo)
	c.Header("Profile-Update-Interval", strconv.Itoa(res.ProfileUpdateInterval))
	c.Header("Profile-Title", res.ProfileTitle)
	c.Header("Content-Disposition", "attachment; filename=\""+res.Filename+"\"")
	c.Header("Cache-Control", "private, no-cache, no-transform")

	c.Data(http.StatusOK, res.ContentType, []byte(res.Content))
}
