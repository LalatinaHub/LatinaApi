package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/LalatinaHub/LatinaApi/pkg/logger"
	"github.com/gin-gonic/gin"
)

// RequestLoggerMiddleware logs request details and attaches request_id / trace_id.
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Extract or generate Request ID / Trace ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			requestID = hex.EncodeToString(b)
		}

		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = requestID
		}

		// Set response headers
		c.Header("X-Request-ID", requestID)
		c.Header("X-Trace-ID", traceID)

		// Set in Gin context
		c.Set("request_id", requestID)
		c.Set("trace_id", traceID)

		c.Next()

		// Skip health ping logs to avoid noise
		if path == "/health" || path == "/api/v1/ping" {
			return
		}

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		logEvent := logger.Info()
		if statusCode >= 500 {
			logEvent = logger.Error()
		} else if statusCode >= 400 {
			logEvent = logger.Warn()
		}

		if raw != "" {
			path = path + "?" + raw
		}

		logEvent.
			Str("request_id", requestID).
			Str("trace_id", traceID).
			Str("client_ip", clientIP).
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Msg("HTTP Request")
	}
}
