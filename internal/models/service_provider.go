package models

import (
	"time"

	"gorm.io/gorm"
)

// ServiceProviderStatus represents the status of a service provider
type ServiceProviderStatus string

const (
	SPStatusPendingApproval ServiceProviderStatus = "pending_approval"
	SPStatusActive          ServiceProviderStatus = "active"
	SPStatusSuspended       ServiceProviderStatus = "suspended"
	SPStatusBanned          ServiceProviderStatus = "banned"
)

// ServiceProviderProfile stores service provider-specific information
type ServiceProviderProfile struct {
	ID     string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID string `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`
	User   *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Business Information
	BusinessName    *string `gorm:"type:varchar(255)" json:"businessName,omitempty"`
	Description     *string `gorm:"type:text" json:"description,omitempty"`
	ServiceCategory string  `gorm:"type:varchar(100);not null" json:"serviceCategory"` // delivery, handyman, general_service
	ServiceType     string  `gorm:"type:varchar(255);not null;index" json:"serviceType"`

	// Verification & Documents
	Status           ServiceProviderStatus `gorm:"type:varchar(50);not null;default:'pending_approval'" json:"status"`
	IsVerified       bool                  `gorm:"default:false" json:"isVerified"`
	VerificationDocs []string              `gorm:"type:jsonb" json:"verificationDocs,omitempty"`

	// Ratings & Performance
	Rating        float64 `gorm:"type:decimal(3,2);default:0" json:"rating"`
	TotalReviews  int     `gorm:"default:0" json:"totalReviews"`
	CompletedJobs int     `gorm:"default:0" json:"completedJobs"`

	// Availability
	IsAvailable  bool     `gorm:"default:true" json:"isAvailable"`
	WorkingHours *string  `gorm:"type:jsonb" json:"workingHours,omitempty"`
	ServiceAreas []string `gorm:"type:jsonb" json:"serviceAreas,omitempty"`

	// Financial
	HourlyRate *float64 `gorm:"type:decimal(10,2)" json:"hourlyRate,omitempty"`
	Currency   string   `gorm:"type:varchar(3);default:'USD'" json:"currency"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ServiceProviderProfile) TableName() string {
	return "service_provider_profiles"
}
