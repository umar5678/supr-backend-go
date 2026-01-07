// internal/modules/fraud/routes.go
package fraud

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	fraud := router.Group("/fraud")
	fraud.Use(authMiddleware) // Admin only - add role check middleware
	{
		fraud.GET("/patterns", handler.ListFraudPatterns)
		fraud.GET("/patterns/:id", handler.GetFraudPattern)
		fraud.POST("/patterns/:id/review", handler.ReviewFraudPattern)
		fraud.GET("/stats", handler.GetFraudStats)
		fraud.GET("/users/:userId/risk-score", handler.CheckUserRiskScore)
	}
}
