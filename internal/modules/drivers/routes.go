// internal/modules/drivers/routes.go
package drivers

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	drivers := router.Group("/drivers")
	drivers.Use(authMiddleware)
	{
		drivers.POST("/register", handler.RegisterDriver)
		drivers.GET("/profile", handler.GetProfile)
		drivers.PUT("/profile", handler.UpdateProfile)
		drivers.PUT("/vehicle", handler.UpdateVehicle)
		drivers.POST("/status", handler.UpdateStatus)
		drivers.POST("/location", handler.UpdateLocation)
		drivers.GET("/wallet", handler.GetWallet)
		drivers.GET("/dashboard", handler.GetDashboard)
		
		// ✅ Wallet management endpoints
		drivers.POST("/wallet/topup", handler.TopUpWallet)
		drivers.GET("/wallet/status", handler.GetWalletStatus)
		drivers.GET("/wallet/transactions", handler.GetWalletTransactionHistory)
	}
}

