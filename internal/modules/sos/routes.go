package sos

import (
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
    sos := router.Group("/sos")
    sos.Use(authMiddleware)
    {
        sos.POST("/trigger", handler.TriggerSOS)
        sos.GET("/active", handler.GetActiveSOS)
        sos.GET("", handler.ListSOS)
        sos.GET("/:id", handler.GetSOS)
        sos.POST("/:id/resolve", handler.ResolveSOS)
        sos.POST("/:id/cancel", handler.CancelSOS)
    }
}
