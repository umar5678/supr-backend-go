package authdto

import (
	"errors"
	"regexp"

	"github.com/umar5678/go-backend/internal/models"
)

// Phone regex pattern (international format)
var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// PhoneSignupRequest for rider/driver signup
type PhoneSignupRequest struct {
	Name  string          `json:"name" binding:"required,min=2,max=255"`
	Phone string          `json:"phone" binding:"required"`
	Role  models.UserRole `json:"role" binding:"required,oneof=rider driver service_provider"`
}

func (r *PhoneSignupRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if len(r.Name) < 2 {
		return errors.New("name must be at least 2 characters")
	}
	if r.Phone == "" {
		return errors.New("phone is required")
	}
	if !phoneRegex.MatchString(r.Phone) {
		return errors.New("invalid phone number format")
	}
	if r.Role != models.RoleRider && r.Role != models.RoleDriver && r.Role != models.RoleServiceProvider {
		return errors.New("role must be either 'rider' or 'driver' or 'service_provider'")
	}
	return nil
}

// PhoneLoginRequest for rider/driver login
type PhoneLoginRequest struct {
	Phone string          `json:"phone" binding:"required"`
	Role  models.UserRole `json:"role" binding:"required,oneof=rider driver service_provider"`
}

func (r *PhoneLoginRequest) Validate() error {
	if r.Phone == "" {
		return errors.New("phone is required")
	}
	if !phoneRegex.MatchString(r.Phone) {
		return errors.New("invalid phone number format")
	}
	if r.Role != models.RoleRider && r.Role != models.RoleDriver && r.Role != models.RoleServiceProvider {
		return errors.New("role must be either 'rider', 'driver' or 'service_provider'")
	}
	return nil
}

// EmailSignupRequest for other roles
type EmailSignupRequest struct {
	Name     string          `json:"name" binding:"required,min=2,max=255"`
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=8"`
	Role     models.UserRole `json:"role" binding:"required"`
}

func (r *EmailSignupRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	// Email-based roles
	validRoles := []models.UserRole{
		models.RoleAdmin,
		models.RoleDeliveryPerson,
		models.RoleServiceProvider,
		models.RoleHandyman,
	}
	isValid := false
	for _, role := range validRoles {
		if r.Role == role {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("invalid role for email signup")
	}
	return nil
}

// EmailLoginRequest for other roles
type EmailLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (r *EmailLoginRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// RefreshTokenRequest
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// UpdateProfileRequest
type UpdateProfileRequest struct {
	Name            *string `json:"name" binding:"omitempty,min=2,max=255"`
	Email           *string `json:"email" binding:"omitempty,email"`
	ProfilePhotoURL *string `json:"profilePhotoUrl" binding:"omitempty,url"`
}
