package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerInfo struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Email   string  `json:"email"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

func (c CustomerInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

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

type BookingInfo struct {
	Day            string    `json:"day"`
	Date           string    `json:"date"`
	Time           string    `json:"time"`
	PreferredTime  time.Time `json:"prefferedTime"`
	QuantityOfPros int       `json:"quantityOfPros" default:"1"`
	ToolsRequired  bool      `json:"toolsRequired"`
	PersonCount    int       `json:"personCount" default:"1"`
	Frequency      *string   `json:"frequency,omitempty"`
}

func (b BookingInfo) Value() (driver.Value, error) {
	return json.Marshal(b)
}

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

type SelectedServiceItem struct {
	ServiceSlug string  `json:"serviceSlug"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

type SelectedServices []SelectedServiceItem

func (s SelectedServices) Value() (driver.Value, error) {
	return json.Marshal(s)
}

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

type SelectedAddonItem struct {
	AddonSlug string  `json:"addonSlug"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type SelectedAddons []SelectedAddonItem

func (s SelectedAddons) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

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

type PaymentInfo struct {
	Method        string  `json:"method"`
	Status        string  `json:"status"`
	Total         float64 `json:"total"`
	AmountPaid    float64 `json:"amountPaid"`
	Voucher       string  `json:"voucher,omitempty"`
	TransactionID string  `json:"transactionId,omitempty"`
}

func (p PaymentInfo) Value() (driver.Value, error) {
	return json.Marshal(p)
}

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

type CancellationInfo struct {
	CancelledBy     string    `json:"cancelledBy"`
	CancelledAt     time.Time `json:"cancelledAt"`
	Reason          string    `json:"reason"`
	CancellationFee float64   `json:"cancellationFee"`
	RefundAmount    float64   `json:"refundAmount"`
}

func (c CancellationInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

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

type ServiceOrderNew struct {
	ID          string `gorm:"type:uuid;primaryKey" json:"id"`
	OrderNumber string `gorm:"type:varchar(50);uniqueIndex;not null" json:"orderNumber"`

	CustomerID   string       `gorm:"type:uuid;not null;index" json:"customerId"`
	Customer     *User        `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	CustomerInfo CustomerInfo `gorm:"type:jsonb;not null" json:"customerInfo"`

	BookingInfo BookingInfo `gorm:"type:jsonb;not null" json:"bookingInfo"`

	CategorySlug     string           `gorm:"type:varchar(255);not null;index" json:"categorySlug"`
	SelectedServices SelectedServices `gorm:"type:jsonb;not null" json:"selectedServices"`
	SelectedAddons   SelectedAddons   `gorm:"type:jsonb" json:"selectedAddons"`
	SpecialNotes     string           `gorm:"type:text" json:"specialNotes"`

	ServicesTotal      float64 `gorm:"type:decimal(10,2);not null" json:"servicesTotal"`
	AddonsTotal        float64 `gorm:"type:decimal(10,2);default:0" json:"addonsTotal"`
	Subtotal           float64 `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	PlatformCommission float64 `gorm:"type:decimal(10,2);not null" json:"platformCommission"`
	TotalPrice         float64 `gorm:"type:decimal(10,2);not null" json:"totalPrice"`

	PaymentInfo  *PaymentInfo `gorm:"type:jsonb" json:"paymentInfo"`
	WalletHoldID *string      `gorm:"type:uuid" json:"walletHoldId,omitempty"`

	AssignedProviderID  *string                 `gorm:"type:uuid;index" json:"assignedProviderId"`
	AssignedProvider    *ServiceProviderProfile `gorm:"foreignKey:AssignedProviderID;references:ID" json:"assignedProvider,omitempty"`
	ProviderAcceptedAt  *time.Time              `json:"providerAcceptedAt"`
	ProviderStartedAt   *time.Time              `json:"providerStartedAt"`
	ProviderCompletedAt *time.Time              `json:"providerCompletedAt"`

	Status string `gorm:"type:varchar(50);not null;default:'pending';index" json:"status"`

	CancellationInfo *CancellationInfo `gorm:"type:jsonb" json:"cancellationInfo,omitempty"`

	CustomerRating  *int       `gorm:"type:int" json:"customerRating"`
	CustomerReview  string     `gorm:"type:text" json:"customerReview"`
	CustomerRatedAt *time.Time `json:"customerRatedAt"`

	ProviderRating  *int       `gorm:"type:int" json:"providerRating"`
	ProviderReview  string     `gorm:"type:text" json:"providerReview"`
	ProviderRatedAt *time.Time `json:"providerRatedAt"`

	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	CompletedAt *time.Time `json:"completedAt"`

	StatusHistory []OrderStatusHistory `gorm:"foreignKey:OrderID" json:"statusHistory,omitempty"`
}

func (o *ServiceOrderNew) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

func (ServiceOrderNew) TableName() string {
	return "service_orders"
}

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

func (o *ServiceOrderNew) CanBeRatedByCustomer() bool {
	return o.Status == "completed" && o.CustomerRating == nil
}

func (o *ServiceOrderNew) CanBeRatedByProvider() bool {
	return o.Status == "completed" && o.ProviderRating == nil
}

func (o *ServiceOrderNew) IsCompleted() bool {
	return o.Status == "completed"
}

func (o *ServiceOrderNew) IsCancelled() bool {
	return o.Status == "cancelled"
}
