package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRejection struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID   string    `gorm:"type:uuid;not null;index:idx_order_provider_rejection,unique" json:"orderId"`
	ProviderID string    `gorm:"type:uuid;not null;index:idx_order_provider_rejection,unique" json:"providerId"`
	Reason    string    `gorm:"type:text" json:"reason"`
	RejectedAt time.Time `gorm:"autoCreateTime;index" json:"rejectedAt"`
}

func (o *OrderRejection) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

func (OrderRejection) TableName() string {
	return "order_rejections"
}
