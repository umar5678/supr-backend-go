package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ServiceNew struct {
	ID                 string         `gorm:"type:uuid;primaryKey" json:"id"`
	Title              string         `gorm:"type:varchar(255);not null" json:"title"`
	LongTitle          string         `gorm:"type:varchar(500)" json:"longTitle"`
	ServiceSlug        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"serviceSlug"`
	CategorySlug       string         `gorm:"type:varchar(255);not null;index" json:"categorySlug"`
	Description        string         `gorm:"type:text" json:"description"`
	LongDescription    string         `gorm:"type:text" json:"longDescription"`
	Highlights         string         `gorm:"type:text" json:"highlights"`
	WhatsIncluded      pq.StringArray `gorm:"type:text[];not null;default:'{}'" json:"whatsIncluded"`
	TermsAndConditions pq.StringArray `gorm:"type:text[]" json:"termsAndConditions"`
	BannerImage        string         `gorm:"type:varchar(500)" json:"bannerImage"`
	Thumbnail          string         `gorm:"type:varchar(500)" json:"thumbnail"`
	Duration           *int           `gorm:"type:int" json:"duration"`
	IsFrequent         bool           `gorm:"default:false" json:"isFrequent"`
	Frequency          string         `gorm:"type:varchar(100)" json:"frequency"`
	SortOrder          int            `gorm:"default:0" json:"sortOrder"`
	IsActive           bool           `gorm:"default:true" json:"isActive"`
	IsAvailable        bool           `gorm:"default:true" json:"isAvailable"`
	BasePrice          *float64       `gorm:"type:decimal(10,2)" json:"basePrice"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *ServiceNew) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

func (ServiceNew) TableName() string {
	return "services"
}

func (s *ServiceNew) IsPublished() bool {
	return s.IsActive && s.IsAvailable && s.DeletedAt.Time.IsZero()
}