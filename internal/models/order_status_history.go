package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StatusHistoryMetadata stores additional context for status changes
type StatusHistoryMetadata map[string]interface{}

// Value implements driver.Valuer for database storage
func (m StatusHistoryMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner for database retrieval
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

// OrderStatusHistory tracks all status changes for audit trail
type OrderStatusHistory struct {
	ID            string                `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID       string                `gorm:"type:uuid;not null;index" json:"orderId"`
	Order         *ServiceOrder         `gorm:"foreignKey:OrderID" json:"-"`
	FromStatus    string                `gorm:"type:varchar(50)" json:"fromStatus"`
	ToStatus      string                `gorm:"type:varchar(50);not null" json:"toStatus"`
	ChangedBy     *string               `gorm:"type:uuid" json:"changedBy"`
	ChangedByUser *User                 `gorm:"foreignKey:ChangedBy" json:"changedByUser,omitempty"`
	ChangedByRole string                `gorm:"type:varchar(50)" json:"changedByRole"` // customer, provider, admin, system
	Notes         string                `gorm:"type:text" json:"notes"`
	Metadata      StatusHistoryMetadata `gorm:"type:jsonb" json:"metadata"`
	CreatedAt     time.Time             `gorm:"autoCreateTime" json:"createdAt"`
}

// BeforeCreate hook to generate UUID
func (h *OrderStatusHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	return nil
}

// TableName specifies the table name
func (OrderStatusHistory) TableName() string {
	return "order_status_history"
}

// NewOrderStatusHistory creates a new status history entry
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
