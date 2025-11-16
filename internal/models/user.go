package models

import (
	"time"

	"gorm.io/gorm"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleRider           UserRole = "rider"
	RoleDriver          UserRole = "driver"
	RoleAdmin           UserRole = "admin"
	RoleDeliveryPerson  UserRole = "delivery_person"
	RoleServiceProvider UserRole = "service_provider"
	RoleHandyman        UserRole = "handyman"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	StatusActive              UserStatus = "active"
	StatusSuspended           UserStatus = "suspended"
	StatusBanned              UserStatus = "banned"
	StatusPendingVerification UserStatus = "pending_verification"
)

// User model
type User struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	Email           *string        `gorm:"type:varchar(255);uniqueIndex" json:"email,omitempty"`
	Phone           *string        `gorm:"type:varchar(20);uniqueIndex" json:"phone,omitempty"`
	Password        *string        `gorm:"type:varchar(255)" json:"-"`
	Role            UserRole       `gorm:"type:user_role;not null;default:'rider'" json:"role"`
	Status          UserStatus     `gorm:"type:user_status;not null;default:'active'" json:"status"`
	ProfilePhotoURL *string        `gorm:"type:varchar(500)" json:"profilePhotoUrl,omitempty"`
	LastLoginAt     *time.Time     `json:"lastLoginAt,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// IsPhoneUser checks if user uses phone authentication
func (u *User) IsPhoneUser() bool {
	return u.Phone != nil && (u.Role == RoleRider || u.Role == RoleDriver)
}

// IsEmailUser checks if user uses email authentication
func (u *User) IsEmailUser() bool {
	return u.Email != nil && u.Password != nil
}

// GetIdentifier returns phone or email for identification
func (u *User) GetIdentifier() string {
	if u.Phone != nil {
		return *u.Phone
	}
	if u.Email != nil {
		return *u.Email
	}
	return ""
}
