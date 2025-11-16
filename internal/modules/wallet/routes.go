package wallet

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	wallet := router.Group("/wallet")
	wallet.Use(authMiddleware)
	{
		// Wallet info
		wallet.GET("", handler.GetWallet)
		wallet.GET("/balance", handler.GetBalance)

		// Funds management
		wallet.POST("/add-funds", handler.AddFunds)
		wallet.POST("/withdraw", handler.WithdrawFunds)
		wallet.POST("/transfer", handler.TransferFunds)

		// Transactions
		wallet.GET("/transactions", handler.ListTransactions)
		wallet.GET("/transactions/:id", handler.GetTransaction)

		// Holds (for internal use by ride system, etc.)
		wallet.POST("/hold", handler.HoldFunds)
		wallet.POST("/hold/release", handler.ReleaseHold)
		wallet.POST("/hold/capture", handler.CaptureHold)
	}
}
