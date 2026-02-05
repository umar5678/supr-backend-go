package profile

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	profile := router.Group("/profile")
	profile.Use(authMiddleware)
	{
		profile.PUT("/emergency-contact", handler.UpdateEmergencyContact)

		profile.POST("/referral/generate", handler.GenerateReferralCode)
		profile.POST("/referral/apply", handler.ApplyReferralCode)
		profile.GET("/referral", handler.GetReferralInfo)

		profile.POST("/kyc", handler.SubmitKYC)
		profile.GET("/kyc", handler.GetKYC)

		profile.POST("/locations", handler.SaveLocation)
		profile.GET("/locations", handler.GetSavedLocations)
		profile.GET("/locations/recent", handler.GetRecentLocations)
		profile.DELETE("/locations/:id", handler.DeleteLocation)
		profile.POST("/locations/:id/default", handler.SetDefaultLocation)
	}
}
