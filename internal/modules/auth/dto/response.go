package authdto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

type AuthResponse struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	User         *UserResponse `json:"user"`
}

type UserResponse struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Email           *string           `json:"email,omitempty"`
	Phone           *string           `json:"phone,omitempty"`
	Role            models.UserRole   `json:"role"`
	Status          models.UserStatus `json:"status"`
	ProfilePhotoURL *string           `json:"profilePhotoUrl,omitempty"`
	LastLoginAt     *time.Time        `json:"lastLoginAt,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
}

func ToUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		Phone:           user.Phone,
		Role:            user.Role,
		Status:          user.Status,
		ProfilePhotoURL: user.ProfilePhotoURL,
		LastLoginAt:     user.LastLoginAt,
		CreatedAt:       user.CreatedAt,
	}
}
