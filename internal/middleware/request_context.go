package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestContext adds request ID and timing to context
func RequestContext(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or extract request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to context
		c.Set("requestID", requestID)
		c.Set("startTime", time.Now())
		c.Set("version", version)

		// Add to response headers
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
