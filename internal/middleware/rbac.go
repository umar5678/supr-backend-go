package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/response"
)

func hasRole(userRole string, allowedRoles []string) bool {
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole(string(models.RoleAdmin))
}

func RequireServiceProvider() gin.HandlerFunc {
	return RequireRole(
		string(models.RoleServiceProvider),
		string(models.RoleHandyman),
		string(models.RoleDeliveryPerson),
	)
}

func RequireRider() gin.HandlerFunc {
	return RequireRole(string(models.RoleRider))
}

func RequireDriver() gin.HandlerFunc {
	return RequireRole(string(models.RoleDriver))
}

func RequireRiderOrDriver() gin.HandlerFunc {
	return RequireRole(
		string(models.RoleRider),
		string(models.RoleDriver),
	)
}

func RequireAdminOrSelf() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.Error(response.UnauthorizedError("Authentication required"))
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.Error(response.UnauthorizedError("Invalid role format"))
			c.Abort()
			return
		}

		if roleStr == string(models.RoleAdmin) {
			c.Next()
			return
		}

		userID, _ := c.Get("userID")
		resourceUserID := c.Param("id")
		if resourceUserID == "" {
			resourceUserID = c.Param("userId")
		}

		if userID.(string) == resourceUserID {
			c.Next()
			return
		}

		c.Error(response.ForbiddenError("You don't have permission to access this resource"))
		c.Abort()
	}
}

func RequireServiceProviderOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.Error(response.UnauthorizedError("Authentication required"))
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.Error(response.UnauthorizedError("Invalid role format"))
			c.Abort()
			return
		}

		if roleStr == string(models.RoleAdmin) {
			c.Next()
			return
		}

		serviceProviderRoles := []string{
			string(models.RoleServiceProvider),
			string(models.RoleHandyman),
			string(models.RoleDeliveryPerson),
		}

		if hasRole(roleStr, serviceProviderRoles) {
			c.Next()
			return
		}

		c.Error(response.ForbiddenError("You don't have permission to access this resource"))
		c.Abort()
	}
}
