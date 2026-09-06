package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns standard CORS configuration allowing common client integrations.
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Authorization", "Content-Type", "X-Request-ID", "X-Trace-ID", "User-Agent",
		},
		ExposeHeaders: []string{
			"Content-Length", "Subscription-Userinfo", "Profile-Update-Interval", "Profile-Title", "Content-Disposition",
		},
		AllowCredentials: true,
	})
}
