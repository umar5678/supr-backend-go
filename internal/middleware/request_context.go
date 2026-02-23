package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestContext(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("requestID", requestID)
		c.Set("startTime", time.Now())
		c.Set("version", version)

		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
