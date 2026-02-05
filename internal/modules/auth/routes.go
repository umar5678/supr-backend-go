package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	auth := router.Group("/auth")
	{
		phone := auth.Group("/phone")
		{
			phone.POST("/signup", handler.PhoneSignup)
			phone.POST("/login", handler.PhoneLogin)
		}

		email := auth.Group("/email")
		{
			email.POST("/signup", handler.EmailSignup)
			email.POST("/login", handler.EmailLogin)
		}

		auth.POST("/refresh", handler.RefreshToken)
		protected := auth.Group("")
		protected.Use(authMiddleware)
		{
			protected.POST("/logout", handler.Logout)
			protected.GET("/profile", handler.GetProfile)
			protected.PUT("/profile", handler.UpdateProfile)
		}
	}
}
