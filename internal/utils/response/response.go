package response

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type Response struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    interface{}   `json:"data,omitempty"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
	Meta    Meta          `json:"meta"`
	Code    string        `json:"code,omitempty"`
}

type Meta struct {
	RequestID  string  `json:"requestId"`
	Timestamp  string  `json:"timestamp"`
	DurationMs float64 `json:"durationMs,omitempty"`
	Path       string  `json:"path"`
	Method     string  `json:"method"`
	Version    string  `json:"version,omitempty"`
}

func Success(c *gin.Context, data interface{}, message string, code ...string) {
	logger.Info("response.Success called", "message", message)
	statusCode := 200
	if c.Request.Method == "POST" {
		statusCode = 201
	}

	logger.Info("building response struct", "statusCode", statusCode)
	resp := Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    extractMeta(c),
	}

	if len(code) > 0 {
		resp.Code = code[0]
	}

	logger.Info("calling c.JSON", "statusCode", statusCode)
	c.JSON(statusCode, resp)
	if c != nil {
		c.Set("responseSent", true)
	}
	logger.Info("c.JSON completed successfully", "statusCode", statusCode)
}

func SendError(c *gin.Context, statusCode int, message string, errors []ErrorDetail, code ...string) {
	resp := Response{
		Success: false,
		Message: message,
		Errors:  errors,
		Meta:    extractMeta(c),
	}

	if len(code) > 0 {
		resp.Code = code[0]
	}

	c.JSON(statusCode, resp)
}

func ValidationError(c *gin.Context, errors []ErrorDetail) {
	SendError(c, 422, "Validation failed", errors, "VALIDATION_ERROR")
}

func NotFound(c *gin.Context, message string) {
	SendError(c, 404, message, nil, "NOT_FOUND")
}

func Unauthorized(c *gin.Context, message string) {
	SendError(c, 401, message, nil, "UNAUTHORIZED")
}

func Forbidden(c *gin.Context, message string) {
	SendError(c, 403, message, nil, "FORBIDDEN")
}

func InternalError(c *gin.Context, message string) {
	SendError(c, 500, message, nil, "INTERNAL_ERROR")
}

func extractMeta(c *gin.Context) Meta {
	requestID, _ := c.Get("requestID")
	startTime, exists := c.Get("startTime")

	meta := Meta{
		RequestID: requestID.(string),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
	}

	if exists {
		duration := time.Since(startTime.(time.Time))
		meta.DurationMs = float64(duration.Microseconds()) / 1000
	}

	version, exists := c.Get("version")
	if exists {
		meta.Version = version.(string)
	}

	return meta
}
