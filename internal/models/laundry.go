package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LaundryServiceCatalog struct {
	ID              string    `gorm:"type:uuid;primaryKey" json:"id"`
	Slug            string    `gorm:"type:varchar(100);uniqueIndex" json:"slug"`
	Title           string    `gorm:"type:varchar(255)" json:"title"`
	Description     string    `gorm:"type:text" json:"description"`
	ColorCode       string    `gorm:"type:varchar(20)" json:"colorCode"`
	BasePrice       float64   `gorm:"type:decimal(10,2)" json:"basePrice"`
	PricingUnit     string    `gorm:"type:varchar(20)" json:"pricingUnit"`
	TurnaroundHours int       `gorm:"default:48" json:"turnaroundHours"`
	ExpressFee      float64   `gorm:"type:decimal(10,2)" json:"expressFee"`
	ExpressHours    int       `gorm:"default:24" json:"expressHours"`
	DisplayOrder    int       `gorm:"default:0" json:"displayOrder"`
	CategorySlug    string    `gorm:"type:varchar(100);default:'laundry';index" json:"categorySlug"`
	IsActive        bool      `gorm:"default:true" json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	Products []LaundryServiceProduct `gorm:"foreignKey:ServiceSlug;references:Slug" json:"products,omitempty"`
}

func (LaundryServiceCatalog) TableName() string {
	return "laundry_service_catalog"
}

type LaundryServiceProduct struct {
	ID                  string    `gorm:"type:uuid;primaryKey" json:"id"`
	ServiceSlug         string    `gorm:"type:varchar(100);not null" json:"serviceSlug"`
	Name                string    `gorm:"type:varchar(255);not null" json:"name"`
	Slug                string    `gorm:"type:varchar(100);not null" json:"slug"`
	Description         string    `gorm:"type:text" json:"description"`
	IconURL             *string   `gorm:"type:varchar(500)" json:"iconUrl,omitempty"`
	Price               *float64  `gorm:"type:decimal(10,2)" json:"price,omitempty"`
	PricingUnit         *string   `gorm:"type:varchar(20)" json:"pricingUnit,omitempty"`
	TypicalWeight       *float64  `gorm:"type:decimal(8,3)" json:"typicalWeight,omitempty"`
	RequiresSpecialCare bool      `gorm:"default:false" json:"requiresSpecialCare"`
	SpecialCareFee      float64   `gorm:"type:decimal(10,2);default:0" json:"specialCareFee"`
	DisplayOrder        int       `gorm:"default:0" json:"displayOrder"`
	CategorySlug        string    `gorm:"type:varchar(100);default:'laundry';index" json:"categorySlug"`
	IsActive            bool      `gorm:"default:true" json:"isActive"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

func (l *LaundryServiceProduct) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

func (LaundryServiceProduct) TableName() string {
	return "laundry_service_products"
}

type LaundryOrderItem struct {
	ID               string     `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID          string     `gorm:"type:uuid;not null" json:"orderId"`
	ServiceSlug      string     `gorm:"type:varchar(100)" json:"serviceSlug"`
	ProductSlug      string     `gorm:"type:varchar(100)" json:"productSlug"`
	ItemType         string     `gorm:"type:varchar(100)" json:"itemType"`
	Quantity         int        `gorm:"default:1" json:"quantity"`
	Weight           *float64   `gorm:"type:decimal(8,3)" json:"weight,omitempty"`
	QRCode           string     `gorm:"type:varchar(255);uniqueIndex" json:"qrCode"`
	Status           string     `gorm:"type:varchar(50);default:'pending'" json:"status"`
	HasIssue         bool       `gorm:"default:false" json:"hasIssue"`
	IssueDescription *string    `gorm:"type:text" json:"issueDescription,omitempty"`
	Price            float64    `gorm:"type:decimal(10,2)" json:"price"`
	ReceivedAt       *time.Time `json:"receivedAt,omitempty"`
	PackedAt         *time.Time `json:"packedAt,omitempty"`
	DeliveredAt      *time.Time `json:"deliveredAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (l *LaundryOrderItem) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	if l.QRCode == "" {
		l.QRCode = "LDY-" + uuid.New().String()[:8]
	}
	return nil
}

func (LaundryOrderItem) TableName() string {
	return "laundry_order_items"
}

type LaundryPickup struct {
	ID          string     `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID     string     `gorm:"type:uuid;uniqueIndex" json:"orderId"`
	ProviderID  *string    `gorm:"type:uuid;index" json:"providerId,omitempty"`
	ScheduledAt time.Time  `gorm:"not null" json:"scheduledAt"`
	ArrivedAt   *time.Time `json:"arrivedAt,omitempty"`
	PickedUpAt  *time.Time `json:"pickedUpAt,omitempty"`
	Status      string     `gorm:"type:varchar(50);default:'scheduled'" json:"status"`
	PhotoURL    *string    `gorm:"type:varchar(500)" json:"photoUrl,omitempty"`
	Notes       string     `gorm:"type:text" json:"notes"`
	BagCount    int        `gorm:"default:0" json:"bagCount"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (p *LaundryPickup) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (LaundryPickup) TableName() string {
	return "laundry_pickups"
}

type LaundryDelivery struct {
	ID                 string     `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID            string     `gorm:"type:uuid;uniqueIndex" json:"orderId"`
	ProviderID         *string    `gorm:"type:uuid;index" json:"providerId,omitempty"`
	ScheduledAt        time.Time  `gorm:"not null" json:"scheduledAt"`
	ArrivedAt          *time.Time `json:"arrivedAt,omitempty"`
	DeliveredAt        *time.Time `json:"deliveredAt,omitempty"`
	Status             string     `gorm:"type:varchar(50);default:'scheduled'" json:"status"`
	PhotoURL           *string    `gorm:"type:varchar(500)" json:"photoUrl,omitempty"`
	RecipientName      *string    `gorm:"type:varchar(255)" json:"recipientName,omitempty"`
	RecipientSignature *string    `gorm:"type:text" json:"recipientSignature,omitempty"`
	Notes              string     `gorm:"type:text" json:"notes"`
	RescheduleCount    int        `gorm:"default:0" json:"rescheduleCount"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

func (d *LaundryDelivery) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

func (LaundryDelivery) TableName() string {
	return "laundry_deliveries"
}

type LaundryIssue struct {
	ID               string     `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID          string     `gorm:"type:uuid;not null" json:"orderId"`
	CustomerID       string     `gorm:"type:uuid;not null" json:"customerId"`
	ProviderID       string     `gorm:"type:uuid;not null" json:"providerId"`
	IssueType        string     `gorm:"type:varchar(100)" json:"issueType"`
	Description      string     `gorm:"type:text" json:"description"`
	Priority         string     `gorm:"type:varchar(20);default:'medium'" json:"priority"`
	Status           string     `gorm:"type:varchar(50);default:'open'" json:"status"`
	Resolution       *string    `gorm:"type:text" json:"resolution,omitempty"`
	RefundAmount     *float64   `gorm:"type:decimal(10,2)" json:"refundAmount,omitempty"`
	CompensationType *string    `gorm:"type:varchar(100)" json:"compensationType,omitempty"`
	ResolvedAt       *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (i *LaundryIssue) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

func (LaundryIssue) TableName() string {
	return "laundry_issues"
}

type LaundryOrder struct {
	ID          string `gorm:"type:uuid;primaryKey" json:"id"`
	OrderNumber string `gorm:"type:varchar(50);uniqueIndex;not null" json:"orderNumber"`

	UserID *string `gorm:"type:uuid;index" json:"userId,omitempty"`

	CategorySlug string `gorm:"type:varchar(255);not null;index" json:"categorySlug"`

	Status    string  `gorm:"type:varchar(50);not null;default:'pending';index" json:"status"`
	Address   string  `gorm:"type:text;not null" json:"address"`
	Latitude  float64 `gorm:"type:decimal(10,8)" json:"lat"`
	Longitude float64 `gorm:"type:decimal(11,8)" json:"lng"`

	PersonCount   int  `gorm:"default:1" json:"personCount"`

	ServiceDate *time.Time `json:"serviceDate,omitempty"`
	Total       float64    `gorm:"type:decimal(10,2);not null" json:"total"`
	Tip         *float64   `gorm:"type:decimal(10,2)" json:"tip,omitempty"`
	IsExpress   bool       `gorm:"type:boolean;default:false" json:"isExpress"`

	ProviderID *string `gorm:"type:uuid;index" json:"providerId,omitempty"`

	CreatedAt time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

func (LaundryOrder) TableName() string {
	return "laundry_orders"
}
