package middleware

import (
	"fmt"
	"sync"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/utils/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type rateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func newRateLimiter(requestsPerSecond, burst int) *rateLimiter {
	return &rateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

func (rl *rateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	if cfg.RequestsPerSecond <= 0 {
		cfg.RequestsPerSecond = 100
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 200
	}

	rl := newRateLimiter(cfg.RequestsPerSecond, cfg.Burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			c.Error(response.TooManyRequests("Rate limit exceeded"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimitByKey(keyFunc func(*gin.Context) string, requestsPerSecond int, burst int) gin.HandlerFunc {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 100
	}
	if burst <= 0 {
		burst = 200
	}

	rl := newRateLimiter(requestsPerSecond, burst)

	return func(c *gin.Context) {
		key := keyFunc(c)
		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			c.Error(response.TooManyRequests(fmt.Sprintf("Rate limit exceeded for key: %s", key)))
			c.Abort()
			return
		}

		c.Next()
	}
}