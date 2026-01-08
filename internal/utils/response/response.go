package response

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// Response represents the standard API response structure
type Response struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    interface{}   `json:"data,omitempty"`
	Errors  []ErrorDetail `json:"errors,omitempty"` // Use ErrorDetail
	Meta    Meta          `json:"meta"`
	Code    string        `json:"code,omitempty"`
}

// Meta contains request metadata
type Meta struct {
	RequestID  string  `json:"requestId"`
	Timestamp  string  `json:"timestamp"`
	DurationMs float64 `json:"durationMs,omitempty"`
	Path       string  `json:"path"`
	Method     string  `json:"method"`
	Version    string  `json:"version,omitempty"`
}

// Success sends a successful response
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
	// Mark in the Gin context that a response has been sent so error middleware
	// doesn't attempt to write another response later (avoids double-write 200->500).
	// Note: use a simple key that middleware can check safely.
	if c != nil {
		c.Set("responseSent", true)
	}
	logger.Info("c.JSON completed successfully", "statusCode", statusCode)
}

// SendError sends an error response
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

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, errors []ErrorDetail) {
	SendError(c, 422, "Validation failed", errors, "VALIDATION_ERROR")
}

// NotFound sends a 404 response
func NotFound(c *gin.Context, message string) {
	SendError(c, 404, message, nil, "NOT_FOUND")
}

// Unauthorized sends a 401 response
func Unauthorized(c *gin.Context, message string) {
	SendError(c, 401, message, nil, "UNAUTHORIZED")
}

// Forbidden sends a 403 response
func Forbidden(c *gin.Context, message string) {
	SendError(c, 403, message, nil, "FORBIDDEN")
}

// InternalError sends a 500 response
func InternalError(c *gin.Context, message string) {
	SendError(c, 500, message, nil, "INTERNAL_ERROR")
}

// extractMeta extracts metadata from the context
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
