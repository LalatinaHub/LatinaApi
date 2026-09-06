package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestLoggerMiddleware_Development(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestLoggerMiddleware(false)) // Development mode

	r.POST("/test", func(c *gin.Context) {
		c.Header("Subscription-Userinfo", "upload=0; download=100; total=500")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 1. Test POST with body in dev mode
	reqBody := `{"vpn":"trojan","mode":"cdn"}`
	req := httptest.NewRequest(http.MethodPost, "/test?format=raw&cc=ID", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TestClient/1.0")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	assert.NotEmpty(t, w.Header().Get("X-Trace-ID"))
	assert.Equal(t, "upload=0; download=100; total=500", w.Header().Get("Subscription-Userinfo"))

	// 2. Test ping route (should pass through without error)
	reqPing := httptest.NewRequest(http.MethodGet, "/ping", nil)
	wPing := httptest.NewRecorder()
	r.ServeHTTP(wPing, reqPing)
	assert.Equal(t, http.StatusOK, wPing.Code)
}

func TestRequestLoggerMiddleware_Production(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(RequestLoggerMiddleware(true)) // Production mode

	r.GET("/prod", func(c *gin.Context) {
		c.String(http.StatusOK, "prod ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/prod", nil)
	req.Header.Set("X-Request-ID", "custom-req-id-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-req-id-123", w.Header().Get("X-Request-ID"))
	assert.Equal(t, "custom-req-id-123", w.Header().Get("X-Trace-ID"))
}
