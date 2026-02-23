package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Addon struct {
	ID                 string         `gorm:"type:uuid;primaryKey" json:"id"`
	Title              string         `gorm:"type:varchar(255);not null" json:"title"`
	AddonSlug          string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"addonSlug"`
	CategorySlug       string         `gorm:"type:varchar(255);not null;index" json:"categorySlug"`
	Description        string         `gorm:"type:text" json:"description"`
	WhatsIncluded      pq.StringArray `gorm:"type:text[]" json:"whatsIncluded"`
	Notes              pq.StringArray `gorm:"type:text[]" json:"notes"`
	Image              string         `gorm:"type:varchar(500)" json:"image"`
	Price              float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	StrikethroughPrice *float64       `gorm:"type:decimal(10,2)" json:"strikethroughPrice"`
	IsActive           bool           `gorm:"default:true" json:"isActive"`
	IsAvailable        bool           `gorm:"default:true" json:"isAvailable"`
	SortOrder          int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *Addon) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

func (Addon) TableName() string {
	return "addons"
}

func (a *Addon) IsPublished() bool {
	return a.IsActive && a.IsAvailable && a.DeletedAt.Time.IsZero()
}

func (a *Addon) HasDiscount() bool {
	return a.StrikethroughPrice != nil && *a.StrikethroughPrice > a.Price
}

func (a *Addon) DiscountPercentage() float64 {
	if !a.HasDiscount() {
		return 0
	}
	return ((*a.StrikethroughPrice - a.Price) / *a.StrikethroughPrice) * 100
}
