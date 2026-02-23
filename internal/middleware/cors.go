package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/config"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if os.Getenv("ENABLE_CORS_MIDDLEWARE") == "false" {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")

		if isAllowedOrigin(origin, cfg.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowedHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		if c.Request.Method != http.MethodOptions {
			if origin != "" && isAllowedOrigin(origin, cfg.AllowedOrigins) {
				c.Header("Access-Control-Allow-Origin", origin)
			}

			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Next()
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}

	for _, o := range allowed {
		if o == "*" {
			return true
		}
		if o == origin {
			return true
		}
	}
	return false
}

func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}
