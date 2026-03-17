package documents

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	documents := router.Group("/documents")
	documents.Use(authMiddleware)
	{
		documents.POST("/upload", handler.UploadDocument)
		documents.GET("", handler.GetDocuments)

		admin := documents.Group("")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("/driver", handler.GetDocumentsByDriver)
			admin.GET("/service-provider", handler.GetDocumentsByServiceProvider)
			admin.GET("/admin", handler.GetDocumentsAdmin)
			admin.POST("/verify", handler.VerifyDocument)
			admin.GET("/pending", handler.GetPendingDocuments)
		}

		documents.DELETE("/:id", handler.DeleteDocument)
	}
}
