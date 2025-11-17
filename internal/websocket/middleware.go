// internal/websocket/middleware.go - UPDATED
package websocket

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/utils/jwt"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// AuthMiddleware authenticates WebSocket connections
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token from query parameter first (for WebSocket connections)
		token := c.Query("token")

		// If not in query, try Authorization header
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication token required",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwt.ValidateToken(token, jwtSecret)
		if err != nil {
			logger.Warn("websocket authentication failed",
				"error", err.Error(),
				"remote_addr", c.Request.RemoteAddr,
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
