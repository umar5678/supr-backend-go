// now
package admin

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

// UserResponse represents a user in admin responses
type UserResponse struct {
	ID              string            `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name            string            `json:"name" example:"John Doe"`
	Email           *string           `json:"email,omitempty" example:"john@example.com"`
	Phone           *string           `json:"phone,omitempty" example:"+1234567890"`
	Role            models.UserRole   `json:"role" example:"service_provider"`
	Status          models.UserStatus `json:"status" example:"active"`
	ProfilePhotoURL *string           `json:"profilePhotoUrl,omitempty" example:"https://example.com/photo.jpg"`
	LastLoginAt     *time.Time        `json:"lastLoginAt,omitempty" example:"2024-01-15T10:30:00Z"`
	CreatedAt       time.Time         `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt       time.Time         `json:"updatedAt" example:"2024-01-15T12:00:00Z"`
}

// ListUsersResponse represents the paginated response for listing users
type ListUsersResponse struct {
	Users []*UserResponse `json:"users"`
	Total int64           `json:"total" example:"100"`
	Page  int             `json:"page" example:"1"`
	Limit int             `json:"limit" example:"20"`
}

// DashboardStatsResponse represents admin dashboard statistics
type DashboardStatsResponse struct {
	TotalUsers    int64                 `json:"totalUsers" example:"1500"`
	UsersByRole   []RoleCountResponse   `json:"usersByRole"`
	UsersByStatus []StatusCountResponse `json:"usersByStatus"`
}

// RoleCountResponse represents user count by role
type RoleCountResponse struct {
	Role  string `json:"role" example:"service_provider"`
	Count int64  `json:"count" example:"250"`
}

// StatusCountResponse represents user count by status
type StatusCountResponse struct {
	Status string `json:"status" example:"active"`
	Count  int64  `json:"count" example:"1200"`
}

// ApproveServiceProviderResponse represents the response after approving a service provider
type ApproveServiceProviderResponse struct {
	Message    string `json:"message" example:"Service provider approved successfully"`
	ProviderID string `json:"providerId" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID     string `json:"userId" example:"660e8400-e29b-41d4-a716-446655440001"`
}

// SuspendUserResponse represents the response after suspending a user
type SuspendUserResponse struct {
	Message string `json:"message" example:"User suspended successfully"`
	UserID  string `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Reason  string `json:"reason" example:"Violation of terms of service"`
}

// UpdateUserStatusResponse represents the response after updating user status
type UpdateUserStatusResponse struct {
	Message   string            `json:"message" example:"User status updated successfully"`
	UserID    string            `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`
	NewStatus models.UserStatus `json:"newStatus" example:"active"`
}
