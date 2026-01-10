package models

import (
	"time"
)

// WalletType represents the type of wallet
type WalletType string

const (
	WalletTypeRider           WalletType = "rider"
	WalletTypeDriver          WalletType = "driver"
	WalletTypePlatform        WalletType = "platform"
	WalletTypeServiceProvider WalletType = "service_provider" // ✅ For handyman, delivery_person, service_provider
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeCredit   TransactionType = "credit"
	TransactionTypeDebit    TransactionType = "debit"
	TransactionTypeRefund   TransactionType = "refund"
	TransactionTypeHold     TransactionType = "hold"
	TransactionTypeRelease  TransactionType = "release"
	TransactionTypeTransfer TransactionType = "transfer"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
	TransactionStatusHeld      TransactionStatus = "held"
	TransactionStatusReleased  TransactionStatus = "released"
)

// Wallet model
type Wallet struct {
	ID              string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID          string     `gorm:"type:uuid;not null;index" json:"userId"`
	WalletType      WalletType `gorm:"type:wallet_type;not null" json:"walletType"`
	Balance         float64    `gorm:"type:decimal(12,2);not null;default:0.00" json:"balance"`
	HeldBalance     float64    `gorm:"type:decimal(12,2);not null;default:0.00" json:"heldBalance"`
	Currency        string     `gorm:"type:varchar(3);not null;default:'INR'" json:"currency"`
	IsActive        bool       `gorm:"not null;default:true" json:"isActive"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	FreeRideCredits float64    `gorm:"type:decimal(12,2);not null;default:0.00" json:"freeRideCredits"`

	// Relations
	User         User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Transactions []WalletTransaction `gorm:"foreignKey:WalletID" json:"transactions,omitempty"`
	Holds        []WalletHold        `gorm:"foreignKey:WalletID" json:"holds,omitempty"`
}

func (Wallet) TableName() string {
	return "wallets"
}

// GetAvailableBalance returns balance minus held balance
func (w *Wallet) GetAvailableBalance() float64 {
	return w.Balance - w.HeldBalance
}

// WalletTransaction model
type WalletTransaction struct {
	ID            string                 `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	WalletID      string                 `gorm:"type:uuid;not null;index" json:"walletId"`
	Type          TransactionType        `gorm:"type:transaction_type;not null" json:"type"`
	Amount        float64                `gorm:"type:decimal(12,2);not null" json:"amount"`
	BalanceBefore float64                `gorm:"type:decimal(12,2);not null" json:"balanceBefore"`
	BalanceAfter  float64                `gorm:"type:decimal(12,2);not null" json:"balanceAfter"`
	Status        TransactionStatus      `gorm:"type:transaction_status;not null;default:'pending'" json:"status"`
	ReferenceType *string                `gorm:"type:varchar(50)" json:"referenceType,omitempty"`
	ReferenceID   *string                `gorm:"type:varchar(50);not null" json:"referenceId"`
	Description   *string                `gorm:"type:text" json:"description,omitempty"`
	PaymentMethod string                 `gorm:"type:varchar(50);not null;default:'credit_card'" json:"paymentMethod"`
	Metadata      map[string]interface{} `gorm:"type:jsonb" json:"metadata,omitempty"`
	ProcessedAt   *time.Time             `json:"processedAt,omitempty"`
	CreatedAt     time.Time              `gorm:"autoCreateTime" json:"createdAt"`

	// Relations
	Wallet Wallet `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
}

func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}

// WalletHold model
type WalletHold struct {
	ID            string            `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	WalletID      string            `gorm:"type:uuid;not null;index" json:"walletId"`
	Amount        float64           `gorm:"type:decimal(12,2);not null" json:"amount"`
	ReferenceType string            `gorm:"type:varchar(50);not null" json:"referenceType"`
	ReferenceID   string            `gorm:"type:uuid;not null" json:"referenceId"`
	Status        TransactionStatus `gorm:"type:transaction_status;not null;default:'held'" json:"status"`
	ExpiresAt     time.Time         `gorm:"not null" json:"expiresAt"`
	ReleasedAt    *time.Time        `json:"releasedAt,omitempty"`
	CreatedAt     time.Time         `gorm:"autoCreateTime" json:"createdAt"`

	// Relations
	Wallet Wallet `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
}

func (WalletHold) TableName() string {
	return "wallet_holds"
}
