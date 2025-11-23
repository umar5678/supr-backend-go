// now
package admin

import "github.com/umar5678/go-backend/internal/models"

// ListUsersQueryParams represents query parameters for listing users
type ListUsersQueryParams struct {
	Role   string `form:"role" example:"service_provider" enums:"rider,driver,admin,delivery_person,service_provider,handyman"`
	Status string `form:"status" example:"active" enums:"active,suspended,banned,pending_verification,pending_approval"`
	Page   string `form:"page" example:"1"`
	Limit  string `form:"limit" example:"20"`
}

// SuspendUserRequest represents the request to suspend a user
type SuspendUserRequest struct {
	Reason string `json:"reason" binding:"required" example:"Violation of terms of service"`
}

// UpdateUserStatusRequest represents the request to update user status
type UpdateUserStatusRequest struct {
	Status models.UserStatus `json:"status" binding:"required" example:"active"`
}

// ApproveServiceProviderParams represents path parameters for approving a service provider
type ApproveServiceProviderParams struct {
	ID string `uri:"id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UserIDParams represents path parameters containing user ID
type UserIDParams struct {
	ID string `uri:"id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}
