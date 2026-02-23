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
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(response.UnauthorizedError("Authorization header required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Error(response.UnauthorizedError("Invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := jwt.ValidateToken(tokenString, cfg.JWT.Secret)
		if err != nil {
			c.Error(response.UnauthorizedError("Invalid or expired token"))
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

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

func OptionalAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		claims, err := jwt.ValidateToken(token, cfg.JWT.Secret)
		if err == nil {
			c.Set("userID", claims.UserID)
			c.Set("userRole", claims.Role)
		}

		c.Next()
	}
}
