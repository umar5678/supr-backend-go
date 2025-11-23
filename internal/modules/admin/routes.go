package admin

// now

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin")
	admin.Use(authMiddleware)            // All admin routes require auth
	admin.Use(middleware.RequireAdmin()) // All routes require admin role
	{
		admin.GET("/users", handler.ListUsers)
		admin.PUT("/users/:id/status", handler.UpdateUserStatus)
		admin.POST("/service-providers/:id/approve", handler.ApproveServiceProvider)
		admin.POST("/users/:id/suspend", handler.SuspendUser)
		admin.GET("/dashboard/stats", handler.GetDashboardStats)
	}
}
