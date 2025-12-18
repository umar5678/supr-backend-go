package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProviderServiceCategory links providers to service categories they can handle
type ProviderServiceCategory struct {
	ID                string `gorm:"type:uuid;primaryKey" json:"id"`
	ProviderID        string `gorm:"type:uuid;not null;index" json:"providerId"`
	Provider          *User  `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	CategorySlug      string `gorm:"type:varchar(255);not null;index" json:"categorySlug"`
	ExpertiseLevel    string `gorm:"type:varchar(50);default:'beginner'" json:"expertiseLevel"` // beginner, intermediate, expert
	YearsOfExperience int    `gorm:"default:0" json:"yearsOfExperience"`
	IsActive          bool   `gorm:"default:true" json:"isActive"`

	// Statistics
	CompletedJobs int     `gorm:"default:0" json:"completedJobs"`
	TotalEarnings float64 `gorm:"type:decimal(12,2);default:0" json:"totalEarnings"`
	AverageRating float64 `gorm:"type:decimal(3,2);default:0" json:"averageRating"`
	TotalRatings  int     `gorm:"default:0" json:"totalRatings"`

	// Timestamps
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// BeforeCreate hook to generate UUID
func (p *ProviderServiceCategory) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// TableName specifies the table name
func (ProviderServiceCategory) TableName() string {
	return "provider_service_categories"
}

// ExpertiseLevels returns valid expertise levels
func ExpertiseLevels() []string {
	return []string{"beginner", "intermediate", "expert"}
}

// IsValidExpertiseLevel checks if the level is valid
func IsValidExpertiseLevel(level string) bool {
	for _, l := range ExpertiseLevels() {
		if l == level {
			return true
		}
	}
	return false
}

// UpdateRating updates the average rating after a new rating
func (p *ProviderServiceCategory) UpdateRating(newRating int) {
	totalPoints := p.AverageRating * float64(p.TotalRatings)
	p.TotalRatings++
	p.AverageRating = (totalPoints + float64(newRating)) / float64(p.TotalRatings)
}

// IncrementCompletedJobs increments job count and adds earnings
func (p *ProviderServiceCategory) IncrementCompletedJobs(earnings float64) {
	p.CompletedJobs++
	p.TotalEarnings += earnings
}
