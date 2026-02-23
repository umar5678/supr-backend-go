package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProviderServiceCategory struct {
	ID                string `gorm:"type:uuid;primaryKey" json:"id"`
	ProviderID        string `gorm:"type:uuid;not null;index" json:"providerId"`
	Provider          *User  `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	CategorySlug      string `gorm:"type:varchar(255);not null;index" json:"categorySlug"`
	ExpertiseLevel    string `gorm:"type:varchar(50);default:'beginner'" json:"expertiseLevel"`
	YearsOfExperience int    `gorm:"default:0" json:"yearsOfExperience"`
	IsActive          bool   `gorm:"default:true" json:"isActive"`

	CompletedJobs int     `gorm:"default:0" json:"completedJobs"`
	TotalEarnings float64 `gorm:"type:decimal(12,2);default:0" json:"totalEarnings"`
	AverageRating float64 `gorm:"type:decimal(3,2);default:0" json:"averageRating"`
	TotalRatings  int     `gorm:"default:0" json:"totalRatings"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (p *ProviderServiceCategory) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (ProviderServiceCategory) TableName() string {
	return "provider_service_categories"
}

func ExpertiseLevels() []string {
	return []string{"beginner", "intermediate", "expert"}
}

func IsValidExpertiseLevel(level string) bool {
	for _, l := range ExpertiseLevels() {
		if l == level {
			return true
		}
	}
	return false
}

func (p *ProviderServiceCategory) UpdateRating(newRating int) {
	totalPoints := p.AverageRating * float64(p.TotalRatings)
	p.TotalRatings++
	p.AverageRating = (totalPoints + float64(newRating)) / float64(p.TotalRatings)
}

func (p *ProviderServiceCategory) IncrementCompletedJobs(earnings float64) {
	p.CompletedJobs++
	p.TotalEarnings += earnings
}
