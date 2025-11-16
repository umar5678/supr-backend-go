package middleware

import (
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"

	"github.com/gin-gonic/gin"
)

// Recovery handles panics and converts them to proper error responses
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("requestID")

				logger.Error("panic recovered",
					"requestID", requestID,
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
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

		// Handle errors set in context
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			requestID, _ := c.Get("requestID")
			logger.Error("request error",
				"requestID", requestID,
				"error", err.Error(),
				"path", c.Request.URL.Path,
			)

			// Check if it's an AppError
			if appErr, ok := err.Err.(*response.AppError); ok {
				if appErr.Internal != nil {
					logger.Error("internal error details",
						"requestID", requestID,
						"internal", appErr.Internal.Error(),
					)
				}
				response.SendError(c, appErr.StatusCode, appErr.Message, appErr.Errors, appErr.Code)
				return
			}

			// Default error response
			response.InternalError(c, err.Error())
		}
	}
}

// ErrorHandler is a simpler error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if response already sent
		if c.Writer.Written() {
			return
		}

		// Check for errors in context
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Handle AppError
			if appErr, ok := err.(*response.AppError); ok {
				response.SendError(c, appErr.StatusCode, appErr.Message, appErr.Errors, appErr.Code)
				return
			}

			// Generic error
			response.InternalError(c, "An error occurred")
		}
	}
}
