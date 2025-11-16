package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		requestID, _ := c.Get("requestID")

		logger.Info("request completed",
			"requestID", requestID,
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", statusCode,
			"duration", duration.Milliseconds(),
			"ip", c.ClientIP(),
			"userAgent", c.Request.UserAgent(),
		)
	}
}
