package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type AppError struct {
	StatusCode int
	Message    string
	Code       string
	Errors     []ErrorDetail
	Internal   error
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

func NewAppError(statusCode int, message, code string, errors []ErrorDetail, internal error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
		Errors:     errors,
		Internal:   internal,
	}
}

func (e *AppError) ToResponse(c *gin.Context) {
	SendError(c, e.StatusCode, e.Message, e.Errors, e.Code)
}

func BadRequest(message string, errors ...ErrorDetail) *AppError {
	if message == "" {
		message = "Bad request"
	}
	return NewAppError(http.StatusBadRequest, message, "BAD_REQUEST", errors, nil)
}

func UnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return NewAppError(http.StatusUnauthorized, message, "UNAUTHORIZED", nil, nil)
}

func ForbiddenError(message string) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return NewAppError(http.StatusForbidden, message, "FORBIDDEN", nil, nil)
}

func NotFoundError(resource string) *AppError {
	message := "Resource not found"
	if resource != "" {
		message = fmt.Sprintf("%s not found", resource)
	}
	return NewAppError(http.StatusNotFound, message, "NOT_FOUND", nil, nil)
}

func ConflictError(message string) *AppError {
	if message == "" {
		message = "Resource conflict"
	}
	return NewAppError(http.StatusConflict, message, "CONFLICT", nil, nil)
}

func NewValidationAppError(message string, errors []ErrorDetail) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return NewAppError(http.StatusUnprocessableEntity, message, "VALIDATION_ERROR", errors, nil)
}

func InternalServerError(message string, internal error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAppError(http.StatusInternalServerError, message, "INTERNAL_ERROR", nil, internal)
}

func TooManyRequests(message string) *AppError {
	if message == "" {
		message = "Too many requests"
	}
	return NewAppError(http.StatusTooManyRequests, message, "RATE_LIMIT_EXCEEDED", nil, nil)
}

func ServiceUnavailable(message string) *AppError {
	if message == "" {
		message = "Service unavailable"
	}
	return NewAppError(http.StatusServiceUnavailable, message, "SERVICE_UNAVAILABLE", nil, nil)
}

func NewValidationErrorDetail(field, message string) ErrorDetail {
	return ErrorDetail{
		Field:   field,
		Message: message,
		Code:    "INVALID_FIELD",
	}
}

func NewErrorDetail(message, code string) ErrorDetail {
	return ErrorDetail{
		Message: message,
		Code:    code,
	}
}
