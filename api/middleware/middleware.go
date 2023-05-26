package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimiter() gin.HandlerFunc {
	var (
		limitRate = limiter.Rate{
			Period: 1 * time.Second,
			Limit:  10,
		}
		limitStore = memory.NewStore()
	)

	return mgin.NewMiddleware(limiter.New(limitStore, limitRate))
}
