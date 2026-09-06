package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type clientBucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter manages token bucket rate limits per client IP.
type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientBucket
	rps     float64
	burst   float64
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientBucket),
		rps:     rps,
		burst:   float64(burst),
	}

	// Periodic cleanup of stale clients
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for ip, b := range rl.clients {
		if b.lastRefill.Before(cutoff) {
			delete(rl.clients, ip)
		}
	}
}

// Middleware returns a Gin middleware enforcing the rate limit.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		b, exists := rl.clients[ip]
		now := time.Now()

		if !exists {
			rl.clients[ip] = &clientBucket{
				tokens:     rl.burst - 1,
				lastRefill: now,
			}
			rl.mu.Unlock()
			c.Next()
			return
		}

		// Refill tokens based on elapsed time
		elapsed := now.Sub(b.lastRefill).Seconds()
		b.tokens += elapsed * rl.rps
		if b.tokens > rl.burst {
			b.tokens = rl.burst
		}
		b.lastRefill = now

		if b.tokens < 1.0 {
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please try again later",
			})
			return
		}

		b.tokens -= 1.0
		rl.mu.Unlock()

		c.Next()
	}
}
