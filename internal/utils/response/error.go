package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorDetail represents validation or field-specific error
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// AppError represents application error with context
type AppError struct {
	StatusCode int
	Message    string
	Code       string
	Errors     []ErrorDetail
	Internal   error // Not exposed to client, used for logging
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// NewAppError creates new application error
func NewAppError(statusCode int, message, code string, errors []ErrorDetail, internal error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
		Errors:     errors,
		Internal:   internal,
	}
}

// ToResponse converts AppError to a Response (for easier integration)
func (e *AppError) ToResponse(c *gin.Context) {
	SendError(c, e.StatusCode, e.Message, e.Errors, e.Code) // Uses SendError func from response.go
}

// Common error constructors

// BadRequest creates 400 error
func BadRequest(message string, errors ...ErrorDetail) *AppError {
	if message == "" {
		message = "Bad request"
	}
	return NewAppError(http.StatusBadRequest, message, "BAD_REQUEST", errors, nil)
}

// UnauthorizedError creates 401 error
func UnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return NewAppError(http.StatusUnauthorized, message, "UNAUTHORIZED", nil, nil)
}

// ForbiddenError creates 403 error
func ForbiddenError(message string) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return NewAppError(http.StatusForbidden, message, "FORBIDDEN", nil, nil)
}

// NotFoundError creates 404 error
func NotFoundError(resource string) *AppError {
	message := "Resource not found"
	if resource != "" {
		message = fmt.Sprintf("%s not found", resource)
	}
	return NewAppError(http.StatusNotFound, message, "NOT_FOUND", nil, nil)
}

// ConflictError creates 409 error
func ConflictError(message string) *AppError {
	if message == "" {
		message = "Resource conflict"
	}
	return NewAppError(http.StatusConflict, message, "CONFLICT", nil, nil)
}

// NewValidationAppError creates 422 error
func NewValidationAppError(message string, errors []ErrorDetail) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return NewAppError(http.StatusUnprocessableEntity, message, "VALIDATION_ERROR", errors, nil)
}

// InternalServerError creates 500 error
func InternalServerError(message string, internal error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAppError(http.StatusInternalServerError, message, "INTERNAL_ERROR", nil, internal)
}

// TooManyRequests creates 429 error
func TooManyRequests(message string) *AppError {
	if message == "" {
		message = "Too many requests"
	}
	return NewAppError(http.StatusTooManyRequests, message, "RATE_LIMIT_EXCEEDED", nil, nil)
}

// ServiceUnavailable creates 503 error
func ServiceUnavailable(message string) *AppError {
	if message == "" {
		message = "Service unavailable"
	}
	return NewAppError(http.StatusServiceUnavailable, message, "SERVICE_UNAVAILABLE", nil, nil)
}

// NewValidationErrorDetail creates single validation error
func NewValidationErrorDetail(field, message string) ErrorDetail {
	return ErrorDetail{
		Field:   field,
		Message: message,
		Code:    "INVALID_FIELD",
	}
}

// NewErrorDetail creates generic error
func NewErrorDetail(message, code string) ErrorDetail {
	return ErrorDetail{
		Message: message,
		Code:    code,
	}
}
