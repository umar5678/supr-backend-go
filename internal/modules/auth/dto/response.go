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
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	Email                 *string           `json:"email,omitempty"`
	Phone                 *string           `json:"phone,omitempty"`
	Gender                *string           `json:"gender,omitempty"`
	Role                  models.UserRole   `json:"role"`
	RidePIN               string            `json:"ridePin"`
	EmergencyContactName  string            `json:"emergencyContactName,omitempty"`
	EmergencyContactPhone string            `json:"emergencyContactPhone,omitempty"`
	ReferralCode          string            `json:"referralCode,omitempty"`
	Status                models.UserStatus `json:"status"`
	ProfilePhotoURL       *string           `json:"profilePhotoUrl,omitempty"`
	LastLoginAt           *time.Time        `json:"lastLoginAt,omitempty"`
	CreatedAt             time.Time         `json:"createdAt"`
}

type DriverResponse struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	Email                 *string           `json:"email,omitempty"`
	Phone                 *string           `json:"phone,omitempty"`
	Gender				*string           `json:"gender,omitempty"`
	Role                  models.UserRole   `json:"role"`
	RidePIN               string            `json:"ridePin"`
	LicenseNumber         string            `json:"licenseNumber,omitempty"`
	EmergencyContactName  string            `json:"emergencyContactName,omitempty"`
	EmergencyContactPhone string            `json:"emergencyContactPhone,omitempty"`
	ReferralCode          string            `json:"referralCode,omitempty"`
	Status                models.UserStatus `json:"status"`
	ProfilePhotoURL       *string           `json:"profilePhotoUrl,omitempty"`
	LastLoginAt           *time.Time        `json:"lastLoginAt,omitempty"`
	CreatedAt             time.Time         `json:"createdAt"`
}

func ToUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:                    user.ID,
		Name:                  user.Name,
		Email:                 user.Email,
		Phone:                 user.Phone,
		Role:                  user.Role,
		Gender:                user.Gender,
		RidePIN:               user.RidePIN,
		EmergencyContactName:  user.EmergencyContactName,
		EmergencyContactPhone: user.EmergencyContactPhone,
		ReferralCode:          user.ReferralCode,
		Status:                user.Status,
		ProfilePhotoURL:       user.ProfilePhotoURL,
		LastLoginAt:           user.LastLoginAt,
		CreatedAt:             user.CreatedAt,
	}
}

func ToDriverResponse(user *models.User, driverProfile *models.DriverProfile) *DriverResponse {
	return &DriverResponse{
		ID:                    user.ID,
		Name:                  user.Name,
		Email:                 user.Email,
		Phone:                 user.Phone,
		Role:				  user.Role,
		LicenseNumber:        driverProfile.LicenseNumber,
		Gender:                user.Gender,
		RidePIN:               user.RidePIN,
		EmergencyContactName:  user.EmergencyContactName,
		EmergencyContactPhone: user.EmergencyContactPhone,
		ReferralCode:          user.ReferralCode,
		Status:                user.Status,
		ProfilePhotoURL:       user.ProfilePhotoURL,
		LastLoginAt:           user.LastLoginAt,
		CreatedAt:             user.CreatedAt,
	}
}