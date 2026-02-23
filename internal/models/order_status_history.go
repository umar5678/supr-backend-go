package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatusHistoryMetadata map[string]interface{}

func (m StatusHistoryMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *StatusHistoryMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = StatusHistoryMetadata{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

type OrderStatusHistory struct {
	ID            string                `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID       string                `gorm:"type:uuid;not null;index" json:"orderId"`
	Order         *ServiceOrder         `gorm:"foreignKey:OrderID" json:"-"`
	FromStatus    string                `gorm:"type:varchar(50)" json:"fromStatus"`
	ToStatus      string                `gorm:"type:varchar(50);not null" json:"toStatus"`
	ChangedBy     *string               `gorm:"type:uuid" json:"changedBy"`
	ChangedByUser *User                 `gorm:"foreignKey:ChangedBy" json:"changedByUser,omitempty"`
	ChangedByRole string                `gorm:"type:varchar(50)" json:"changedByRole"`
	Notes         string                `gorm:"type:text" json:"notes"`
	Metadata      StatusHistoryMetadata `gorm:"type:jsonb" json:"metadata"`
	CreatedAt     time.Time             `gorm:"autoCreateTime" json:"createdAt"`
}

func (h *OrderStatusHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	return nil
}

func (OrderStatusHistory) TableName() string {
	return "order_status_history"
}

func NewOrderStatusHistory(orderID, fromStatus, toStatus string, changedBy *string, role, notes string, metadata StatusHistoryMetadata) *OrderStatusHistory {
	return &OrderStatusHistory{
		OrderID:       orderID,
		FromStatus:    fromStatus,
		ToStatus:      toStatus,
		ChangedBy:     changedBy,
		ChangedByRole: role,
		Notes:         notes,
		Metadata:      metadata,
	}
}
