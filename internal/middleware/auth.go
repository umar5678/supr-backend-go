package middleware

import (
	"strings"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/utils/jwt"
	"github.com/umar5678/go-backend/internal/utils/response"

	"github.com/gin-gonic/gin"
)

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(response.UnauthorizedError("Authorization header required"))
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Error(response.UnauthorizedError("Invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwt.ValidateToken(tokenString, cfg.JWT.Secret)
		if err != nil {
			c.Error(response.UnauthorizedError("Invalid or expired token"))
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.Error(response.ForbiddenError("Insufficient permissions"))
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		c.Error(response.ForbiddenError("Insufficient permissions"))
		c.Abort()
	}
}
