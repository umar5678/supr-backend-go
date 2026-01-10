package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/config"
)

// CORS middleware that can be disabled when behind nginx proxy
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if CORS middleware should be disabled (when behind nginx)
		if os.Getenv("ENABLE_CORS_MIDDLEWARE") == "false" {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		if isAllowedOrigin(origin, cfg.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// Always set these headers for CORS
		c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowedHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle credentials
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight (OPTIONS) requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Add headers to actual requests
		if c.Request.Method != http.MethodOptions {
			// Ensure origin header is set for non-preflight requests too
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

// isAllowedOrigin checks if the given origin is in the allowed list
func isAllowedOrigin(origin string, allowed []string) bool {
	// Empty origin is allowed (for CLI tools like k6, curl, Postman without browser)
	// Only reject if we have a specific origin whitelist and it's not in there
	if origin == "" {
		return false // Allow requests without Origin header (CLI tools, server-to-server)
	}

	for _, o := range allowed {
		// Allow all origins if "*" is in the list (not recommended for production)
		if o == "*" {
			return true
		}
		// Exact match
		if o == origin {
			return true
		}
	}
	return false
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// package middleware

// import (
// 	"net/http"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"github.com/umar5678/go-backend/internal/config"
// )

// func CORS(cfg config.CORSConfig) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		origin := c.GetHeader("Origin")

// 		// Check if origin is allowed
// 		if isAllowedOrigin(origin, cfg.AllowedOrigins) {
// 			c.Header("Access-Control-Allow-Origin", origin)
// 		}

// 		// Always set these headers for CORS
// 		c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))
// 		c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowedHeaders, ", "))
// 		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range, Authorization")
// 		c.Header("Access-Control-Max-Age", "86400")

// 		// Handle credentials
// 		if cfg.AllowCredentials {
// 			c.Header("Access-Control-Allow-Credentials", "true")
// 		}

// 		// Handle preflight (OPTIONS) requests
// 		if c.Request.Method == http.MethodOptions {
// 			c.AbortWithStatus(http.StatusNoContent)
// 			return
// 		}

// 		// Add headers to actual requests
// 		if c.Request.Method != http.MethodOptions {
// 			// Ensure origin header is set for non-preflight requests too
// 			if origin != "" && isAllowedOrigin(origin, cfg.AllowedOrigins) {
// 				c.Header("Access-Control-Allow-Origin", origin)
// 			}

// 			if cfg.AllowCredentials {
// 				c.Header("Access-Control-Allow-Credentials", "true")
// 			}
// 		}

// 		c.Next()
// 	}
// }

// // isAllowedOrigin checks if the given origin is in the allowed list
// func isAllowedOrigin(origin string, allowed []string) bool {
// 	// Empty origin is not allowed
// 	if origin == "" {
// 		return false
// 	}

// 	for _, o := range allowed {
// 		// Allow all origins if "*" is in the list (not recommended for production)
// 		if o == "*" {
// 			return true
// 		}
// 		// Exact match
// 		if o == origin {
// 			return true
// 		}
// 	}
// 	return false
// }

// // joinStrings joins a slice of strings with a separator
// func joinStrings(strs []string, sep string) string {
// 	return strings.Join(strs, sep)
// }

// package middleware

// import (
// 	"github.com/gin-gonic/gin"
// 	"github.com/umar5678/go-backend/internal/config"
// )

// func CORS(cfg config.CORSConfig) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Set CORS headers
// 		origin := c.GetHeader("Origin")
// 		if isAllowedOrigin(origin, cfg.AllowedOrigins) {
// 			c.Header("Access-Control-Allow-Origin", origin)
// 		}

// 		c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))
// 		c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowedHeaders, ", "))

// 		if cfg.AllowCredentials {
// 			c.Header("Access-Control-Allow-Credentials", "true")
// 		}

// 		// Handle preflight
// 		if c.Request.Method == "OPTIONS" {
// 			c.AbortWithStatus(204)
// 			return
// 		}

// 		c.Next()
// 	}
// }

// func isAllowedOrigin(origin string, allowed []string) bool {
// 	for _, o := range allowed {
// 		if o == "*" || o == origin {
// 			return true
// 		}
// 	}
// 	return false
// }

// func joinStrings(strs []string, sep string) string {
// 	result := ""
// 	for i, s := range strs {
// 		if i > 0 {
// 			result += sep
// 		}
// 		result += s
// 	}
// 	return result
// }
