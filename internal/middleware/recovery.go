package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

// Recovery handles panics and converts them to proper error responses
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("requestID")

				logger.Error("panic recovered in Recovery middleware",
					"requestID", requestID,
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"stack", fmt.Sprintf("%v", err),
				)

				// Check if it's an AppError
				if appErr, ok := err.(*response.AppError); ok {
					response.SendError(c, appErr.StatusCode, appErr.Message, appErr.Errors, appErr.Code)
					return
				}

				// Generic error
				response.InternalError(c, "An unexpected error occurred")
				c.Abort()
			}
		}()

		c.Next()

		// REMOVE THIS ENTIRE BLOCK - Move error handling to ErrorHandler
		// if len(c.Errors) > 0 { ... }
	}
}

// package middleware

// import (
// 	"github.com/umar5678/go-backend/internal/utils/logger"
// 	"github.com/umar5678/go-backend/internal/utils/response"

// 	"github.com/gin-gonic/gin"
// )

// // Recovery handles panics and converts them to proper error responses
// func Recovery() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		defer func() {
// 			if err := recover(); err != nil {
// 				requestID, _ := c.Get("requestID")

// 				logger.Error("panic recovered",
// 					"requestID", requestID,
// 					"error", err,
// 					"path", c.Request.URL.Path,
// 					"method", c.Request.Method,
// 				)

// 				// Check if it's an AppError
// 				if appErr, ok := err.(*response.AppError); ok {
// 					response.SendError(c, appErr.StatusCode, appErr.Message, appErr.Errors, appErr.Code)
// 					return
// 				}

// 				// Generic error
// 				response.InternalError(c, "An unexpected error occurred")
// 				c.Abort()
// 			}
// 		}()

// 		c.Next()

// 		// Handle errors set in context
// 		if len(c.Errors) > 0 {
// 			err := c.Errors.Last()

// 			requestID, _ := c.Get("requestID")
// 			logger.Error("request error",
// 				"requestID", requestID,
// 				"error", err.Error(),
// 				"path", c.Request.URL.Path,
// 			)

// 			// Check if it's an AppError
// 			if appErr, ok := err.Err.(*response.AppError); ok {
// 				if appErr.Internal != nil {
// 					logger.Error("internal error details",
// 						"requestID", requestID,
// 						"internal", appErr.Internal.Error(),
// 					)
// 				}
// 				response.SendError(c, appErr.StatusCode, appErr.Message, appErr.Errors, appErr.Code)
// 				return
// 			}

// 			// Default error response
// 			response.InternalError(c, err.Error())
// 		}
// 	}
// }

// // ErrorHandler is a simpler error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("ErrorHandler: before c.Next()", "path", c.Request.URL.Path, "method", c.Request.Method)
		c.Next()
		logger.Info("ErrorHandler: after c.Next()", "path", c.Request.URL.Path, "method", c.Request.Method, "written", c.Writer.Written(), "statusCode", c.Writer.Status())

		// Check if response already sent via Gin Writer or explicit flag set by response.Success
		if c.Writer.Written() {
			logger.Info("ErrorHandler: response already written (writer), skipping error handling", "statusCode", c.Writer.Status())
			return
		}
		if sent, exists := c.Get("responseSent"); exists {
			if b, ok := sent.(bool); ok && b {
				logger.Info("ErrorHandler: responseSent flag detected, skipping error handling", "statusCode", c.Writer.Status())
				return
			}
		}

		// Check for errors in context
		if len(c.Errors) > 0 {
			logger.Error("ErrorHandler: errors found in context", "errorCount", len(c.Errors), "lastError", c.Errors.Last().Err)
			err := c.Errors.Last().Err

			// Handle AppError
			if appErr, ok := err.(*response.AppError); ok {
				logger.Info("ErrorHandler: AppError detected, sending error response", "statusCode", appErr.StatusCode, "message", appErr.Message)
				response.SendError(c, appErr.StatusCode, appErr.Message, appErr.Errors, appErr.Code)
				return
			}

			// Generic error
			logger.Info("ErrorHandler: generic error, sending internal error response")
			response.InternalError(c, "An error occurred")
		}
	}
}
