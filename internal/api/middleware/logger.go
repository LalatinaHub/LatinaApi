package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/LalatinaHub/LatinaApi/pkg/logger"
	"github.com/gin-gonic/gin"
)

const maxDevBodyLog = 2048 // 2 KB max capture for development logs

// responseBodyWriter wraps gin.ResponseWriter to capture response body in development mode.
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	if w.body != nil && w.body.Len() < maxDevBodyLog {
		remaining := maxDevBodyLog - w.body.Len()
		if len(b) > remaining {
			w.body.Write(b[:remaining])
		} else {
			w.body.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *responseBodyWriter) WriteString(s string) (int, error) {
	if w.body != nil && w.body.Len() < maxDevBodyLog {
		remaining := maxDevBodyLog - w.body.Len()
		if len(s) > remaining {
			w.body.WriteString(s[:remaining])
		} else {
			w.body.WriteString(s)
		}
	}
	return w.ResponseWriter.WriteString(s)
}

// RequestLoggerMiddleware logs request details and attaches request_id / trace_id.
// In development mode, it provides enriched IN/OUT request and response logging with headers, query parameters, and body payloads.
func RequestLoggerMiddleware(isProduction ...bool) gin.HandlerFunc {
	isProd := false
	if len(isProduction) > 0 {
		isProd = isProduction[0]
	} else if gin.Mode() == gin.ReleaseMode {
		isProd = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// Skip health check and ping polling endpoints to eliminate noise
		if path == "/health" || path == "/ping" || path == "/api/v1/ping" {
			c.Next()
			return
		}

		// Extract or generate unique Request ID / Trace ID
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

		// Attach IDs to response and context
		c.Header("X-Request-ID", requestID)
		c.Header("X-Trace-ID", traceID)
		c.Set("request_id", requestID)
		c.Set("trace_id", traceID)

		clientIP := c.ClientIP()
		method := c.Request.Method

		var (
			wb      *responseBodyWriter
			reqBody []byte
		)

		if !isProd {
			// In development mode, capture incoming request body
			if c.Request.Body != nil && (method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch) {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil {
					reqBody = bodyBytes
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			// Log incoming HTTP request (IN)
			inEvent := logger.Info().
				Str("direction", "IN").
				Str("request_id", requestID).
				Str("method", method).
				Str("path", path).
				Str("client_ip", clientIP)

			if rawQuery != "" {
				inEvent.Str("query", rawQuery)
			}
			if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
				inEvent.Str("user_agent", userAgent)
			}
			if contentType := c.GetHeader("Content-Type"); contentType != "" {
				inEvent.Str("content_type", contentType)
			}
			if len(reqBody) > 0 {
				bodyStr := string(reqBody)
				if len(bodyStr) > maxDevBodyLog {
					bodyStr = bodyStr[:maxDevBodyLog] + "... [truncated]"
				}
				inEvent.Str("body", bodyStr)
			}
			inEvent.Msg("--> HTTP IN")

			// Wrap ResponseWriter to capture outgoing response body
			wb = &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:           bytes.NewBuffer(nil),
			}
			c.Writer = wb
		}

		// Process request through handler pipeline
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logEvent := logger.Info()
		if statusCode >= 500 {
			logEvent = logger.Error()
		} else if statusCode >= 400 {
			logEvent = logger.Warn()
		}

		fullPath := path
		if rawQuery != "" {
			fullPath = path + "?" + rawQuery
		}

		if !isProd {
			// Enriched outgoing response log (OUT) in development mode
			outEvent := logEvent.
				Str("direction", "OUT").
				Str("request_id", requestID).
				Str("method", method).
				Str("path", fullPath).
				Int("status", statusCode).
				Dur("latency", latency).
				Int("size_bytes", c.Writer.Size())

			if respContentType := c.Writer.Header().Get("Content-Type"); respContentType != "" {
				outEvent.Str("content_type", respContentType)
			}
			if subUserInfo := c.Writer.Header().Get("Subscription-Userinfo"); subUserInfo != "" {
				outEvent.Str("subscription_userinfo", subUserInfo)
			}
			if contentDisp := c.Writer.Header().Get("Content-Disposition"); contentDisp != "" {
				outEvent.Str("content_disposition", contentDisp)
			}
			if profileTitle := c.Writer.Header().Get("Profile-Title"); profileTitle != "" {
				outEvent.Str("profile_title", profileTitle)
			}

			// Capture response body preview
			if wb != nil && wb.body != nil && wb.body.Len() > 0 {
				respStr := wb.body.String()
				if len(respStr) > maxDevBodyLog {
					respStr = respStr[:maxDevBodyLog] + "... [truncated]"
				}
				outEvent.Str("response_body", respStr)
			}

			if len(c.Errors) > 0 {
				outEvent.Str("errors", c.Errors.String())
			}

			outEvent.Msg("<-- HTTP OUT")
		} else {
			// Production mode: high-throughput single-line structured JSON log
			logEvent.
				Str("request_id", requestID).
				Str("trace_id", traceID).
				Str("client_ip", clientIP).
				Str("method", method).
				Str("path", fullPath).
				Int("status", statusCode).
				Dur("latency", latency).
				Int("size_bytes", c.Writer.Size()).
				Msg("HTTP Request")
		}
	}
}
