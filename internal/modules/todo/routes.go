package todos

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	todos := router.Group("/todos")
	todos.Use(authMiddleware) // All todo routes require authentication
	{
		todos.POST("", handler.Create)
		todos.GET("", handler.GetAll)
		todos.GET("/:id", handler.GetByID)
		todos.PUT("/:id", handler.Update)
		todos.DELETE("/:id", handler.Delete)
	}
}