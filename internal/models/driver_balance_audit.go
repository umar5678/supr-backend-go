package models

import "time"

// DriverBalanceAudit tracks all driver wallet balance changes
type DriverBalanceAudit struct {
	ID                   string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	DriverID             string    `gorm:"type:uuid;not null;index" json:"driverId"`
	UserID               string    `gorm:"type:uuid;not null;index" json:"userId"`
	PreviousBalance      float64   `gorm:"type:decimal(10,2);not null" json:"previousBalance"`
	NewBalance           float64   `gorm:"type:decimal(10,2);not null" json:"newBalance"`
	ChangeAmount         float64   `gorm:"type:decimal(10,2);not null" json:"changeAmount"`
	Action               string    `gorm:"type:varchar(100);not null;index" json:"action"` // commission_deducted, penalty_applied, topup_added, etc.
	Reason               *string   `gorm:"type:varchar(255)" json:"reason,omitempty"`
	TriggeredRestriction bool      `gorm:"default:false;index" json:"triggeredRestriction"` // Whether this action triggered account restriction
	CreatedAt            time.Time `gorm:"autoCreateTime;index" json:"createdAt"`

	// Relations
	DriverProfile DriverProfile `gorm:"foreignKey:DriverID" json:"driverProfile,omitempty"`
	User          User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (DriverBalanceAudit) TableName() string {
	return "driver_balance_audit"
}
