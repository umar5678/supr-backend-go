package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomerInfo stores customer details snapshot at order time
type CustomerInfo struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Email   string  `json:"email"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

// Value implements driver.Valuer for database storage
func (c CustomerInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner for database retrieval
func (c *CustomerInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

// BookingInfo stores booking date/time information
type BookingInfo struct {
	Day            string    `json:"day"`                        // e.g., "Monday"
	Date           string    `json:"date"`                       // YYYY-MM-DD
	Time           string    `json:"time"`                       // HH:MM
	PreferredTime  time.Time `json:"prefferedTime"`              // HH:MM
	QuantityOfPros int       `json:"quantityOfPros" default:"1"` // number of service providers
}

// Value implements driver.Valuer for database storage
func (b BookingInfo) Value() (driver.Value, error) {
	return json.Marshal(b)
}

// Scan implements sql.Scanner for database retrieval
func (b *BookingInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, b)
}

// SelectedServiceItem represents a service in the order
type SelectedServiceItem struct {
	ServiceSlug string  `json:"serviceSlug"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

// SelectedServices is a slice of selected service items
type SelectedServices []SelectedServiceItem

// Value implements driver.Valuer for database storage
func (s SelectedServices) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for database retrieval
func (s *SelectedServices) Scan(value interface{}) error {
	if value == nil {
		*s = SelectedServices{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// SelectedAddonItem represents an addon in the order
type SelectedAddonItem struct {
	AddonSlug string  `json:"addonSlug"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// SelectedAddons is a slice of selected addon items
type SelectedAddons []SelectedAddonItem

// Value implements driver.Valuer for database storage
func (s SelectedAddons) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner for database retrieval
func (s *SelectedAddons) Scan(value interface{}) error {
	if value == nil {
		*s = SelectedAddons{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// PaymentInfo stores payment details
type PaymentInfo struct {
	Method        string  `json:"method"` // wallet, cash, card
	Status        string  `json:"status"` // pending, completed, failed, refunded
	Total         float64 `json:"total"`
	AmountPaid    float64 `json:"amountPaid"`
	Voucher       string  `json:"voucher,omitempty"`
	TransactionID string  `json:"transactionId,omitempty"`
}

// Value implements driver.Valuer for database storage
func (p PaymentInfo) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for database retrieval
func (p *PaymentInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

// CancellationInfo stores cancellation details
type CancellationInfo struct {
	CancelledBy     string    `json:"cancelledBy"` // customer, provider, admin, system
	CancelledAt     time.Time `json:"cancelledAt"`
	Reason          string    `json:"reason"`
	CancellationFee float64   `json:"cancellationFee"`
	RefundAmount    float64   `json:"refundAmount"`
}

// Value implements driver.Valuer for database storage
func (c CancellationInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner for database retrieval
func (c *CancellationInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

// ServiceOrder represents a home service booking
type ServiceOrderNew struct {
	ID          string `gorm:"type:uuid;primaryKey" json:"id"`
	OrderNumber string `gorm:"type:varchar(50);uniqueIndex;not null" json:"orderNumber"`

	// Customer
	CustomerID   string       `gorm:"type:uuid;not null;index" json:"customerId"`
	Customer     *User        `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	CustomerInfo CustomerInfo `gorm:"type:jsonb;not null" json:"customerInfo"`

	// Booking
	BookingInfo BookingInfo `gorm:"type:jsonb;not null" json:"bookingInfo"`

	// Services
	CategorySlug     string           `gorm:"type:varchar(255);not null;index" json:"categorySlug"`
	SelectedServices SelectedServices `gorm:"type:jsonb;not null" json:"selectedServices"`
	SelectedAddons   SelectedAddons   `gorm:"type:jsonb" json:"selectedAddons"`
	SpecialNotes     string           `gorm:"type:text" json:"specialNotes"`

	// Pricing
	ServicesTotal      float64 `gorm:"type:decimal(10,2);not null" json:"servicesTotal"`
	AddonsTotal        float64 `gorm:"type:decimal(10,2);default:0" json:"addonsTotal"`
	Subtotal           float64 `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	PlatformCommission float64 `gorm:"type:decimal(10,2);not null" json:"platformCommission"`
	TotalPrice         float64 `gorm:"type:decimal(10,2);not null" json:"totalPrice"`

	// Payment
	PaymentInfo  *PaymentInfo `gorm:"type:jsonb" json:"paymentInfo"`
	WalletHoldID *string      `gorm:"type:uuid" json:"walletHoldId,omitempty"`

	// Provider
	AssignedProviderID  *string                 `gorm:"type:uuid;index" json:"assignedProviderId"`
	AssignedProvider    *ServiceProviderProfile `gorm:"foreignKey:AssignedProviderID;references:ID" json:"assignedProvider,omitempty"`
	ProviderAcceptedAt  *time.Time              `json:"providerAcceptedAt"`
	ProviderStartedAt   *time.Time              `json:"providerStartedAt"`
	ProviderCompletedAt *time.Time              `json:"providerCompletedAt"`

	// Status
	Status string `gorm:"type:varchar(50);not null;default:'pending';index" json:"status"`

	// Cancellation
	CancellationInfo *CancellationInfo `gorm:"type:jsonb" json:"cancellationInfo,omitempty"`

	// Customer Rating (of Provider)
	CustomerRating  *int       `gorm:"type:int" json:"customerRating"`
	CustomerReview  string     `gorm:"type:text" json:"customerReview"`
	CustomerRatedAt *time.Time `json:"customerRatedAt"`

	// Provider Rating (of Customer)
	ProviderRating  *int       `gorm:"type:int" json:"providerRating"`
	ProviderReview  string     `gorm:"type:text" json:"providerReview"`
	ProviderRatedAt *time.Time `json:"providerRatedAt"`

	// Timestamps
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	CompletedAt *time.Time `json:"completedAt"`

	// Status History (loaded separately)
	StatusHistory []OrderStatusHistory `gorm:"foreignKey:OrderID" json:"statusHistory,omitempty"`
}

// BeforeCreate hook to generate UUID
func (o *ServiceOrderNew) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// TableName specifies the table name
func (ServiceOrderNew) TableName() string {
	return "service_orders"
}

// CanBeCancelled checks if order can be cancelled
func (o *ServiceOrderNew) CanBeCancelled() bool {
	cancelableStatuses := []string{
		"pending",
		"searching_provider",
		"assigned",
		"accepted",
	}
	for _, s := range cancelableStatuses {
		if o.Status == s {
			return true
		}
	}
	return false
}

// CanBeRatedByCustomer checks if customer can rate the order
func (o *ServiceOrderNew) CanBeRatedByCustomer() bool {
	return o.Status == "completed" && o.CustomerRating == nil
}

// CanBeRatedByProvider checks if provider can rate the order
func (o *ServiceOrderNew) CanBeRatedByProvider() bool {
	return o.Status == "completed" && o.ProviderRating == nil
}

// IsCompleted checks if order is completed
func (o *ServiceOrder) IsCompleted() bool {
	return o.Status == "completed"
}

// IsCancelled checks if order is cancelled
func (o *ServiceOrder) IsCancelled() bool {
	return o.Status == "cancelled"
}
