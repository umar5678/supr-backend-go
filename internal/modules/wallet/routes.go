package wallet

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	wallet := router.Group("/wallet")
	wallet.Use(authMiddleware)
	{
		wallet.GET("", handler.GetWallet)
		wallet.GET("/balance", handler.GetBalance)

		wallet.POST("/add-funds", handler.AddFunds)
		wallet.POST("/withdraw", handler.WithdrawFunds)
		wallet.POST("/transfer", handler.TransferFunds)

		wallet.POST("/hold", handler.HoldFunds)
		wallet.POST("/hold/release", handler.ReleaseHold)
		wallet.POST("/hold/capture", handler.CaptureHold)

		wallet.GET("/transactions", handler.GetTransactionHistory)
		wallet.GET("/transactions/:id", handler.GetTransaction)

		wallet.POST("/cash/collect", middleware.RequireRole("driver"), handler.RecordCashCollection)
		wallet.POST("/cash/settle", middleware.RequireRole("driver"), handler.RecordCashPayment)
	}
}
