package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/config"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		origin := c.GetHeader("Origin")
		if isAllowedOrigin(origin, cfg.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowedHeaders, ", "))

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	for _, o := range allowed {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
