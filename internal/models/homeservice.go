package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// JSONBMap for storing flexible JSON data in PostgreSQL
type JSONBMap map[string]interface{}

func (j JSONBMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONBMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// ServiceCategory represents main categories (Women's Salon, Men's Salon, etc.)
type ServiceCategory struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IconURL     string         `gorm:"type:varchar(500)" json:"iconUrl"`
	BannerImage string         `gorm:"type:varchar(500)" json:"bannerImage"`
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	SortOrder   int            `gorm:"default:0" json:"sortOrder"`
	Highlights  pq.StringArray `gorm:"type:text[];default:'{}'" json:"highlights"` // PostgreSQL array
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	Tabs     []ServiceTab   `gorm:"foreignKey:CategoryID" json:"tabs,omitempty"`
	Services []Service      `gorm:"foreignKey:CategoryID" json:"services,omitempty"`
	AddOns   []AddOnService `gorm:"foreignKey:CategoryID" json:"addOns,omitempty"`
}

func (ServiceCategory) TableName() string { return "service_categories" }

// ServiceTab represents subcategories/tabs (Bestsellers, Hair, Nails, etc.)
type ServiceTab struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CategoryID  uint      `gorm:"not null;index" json:"categoryId"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IconURL     string    `gorm:"type:varchar(500)" json:"iconUrl"`
	BannerTitle string    `gorm:"type:varchar(255)" json:"bannerTitle"`
	BannerDesc  string    `gorm:"type:varchar(500)" json:"bannerDescription"`
	BannerImage string    `gorm:"type:varchar(500)" json:"bannerImage"`
	IsActive    bool      `gorm:"default:true" json:"isActive"`
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	Category *ServiceCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (ServiceTab) TableName() string { return "service_tabs" }

// Service represents individual services
type Service struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	CategoryID          uint      `gorm:"not null;index" json:"categoryId"`
	TabID               uint      `gorm:"not null;index" json:"tabId"` // Which tab this belongs to
	Name                string    `gorm:"type:varchar(255);not null" json:"name"`
	Description         string    `gorm:"type:text;not null" json:"description"`
	ImageURL            string    `gorm:"type:varchar(500)" json:"imageUrl"`
	BasePrice           float64   `gorm:"type:decimal(10,2);not null" json:"basePrice"`
	OriginalPrice       float64   `gorm:"type:decimal(10,2)" json:"originalPrice"`
	DiscountPercentage  int       `gorm:"default:0" json:"discountPercentage"`
	PricingModel        string    `gorm:"type:varchar(50);not null;default:'fixed'" json:"pricingModel"` // fixed, hourly, per_unit
	BaseDurationMinutes int       `gorm:"not null" json:"baseDurationMinutes"`
	MaxQuantity         int       `gorm:"default:1" json:"maxQuantity"`
	IsActive            bool      `gorm:"default:true" json:"isActive"`
	IsFeatured          bool      `gorm:"default:false" json:"isFeatured"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	Category *ServiceCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tab      *ServiceTab      `gorm:"foreignKey:TabID" json:"tab,omitempty"`
	Options  []ServiceOption  `gorm:"foreignKey:ServiceID" json:"options,omitempty"`
}

func (Service) TableName() string { return "services" }

// AddOnService represents optional add-ons
type AddOnService struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CategoryID      uint      `gorm:"not null;index" json:"categoryId"`
	Title           string    `gorm:"type:varchar(255);not null" json:"title"`
	Description     string    `gorm:"type:text" json:"description"`
	ImageURL        string    `gorm:"type:varchar(500)" json:"imageUrl"`
	Price           float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	OriginalPrice   float64   `gorm:"type:decimal(10,2)" json:"originalPrice"`
	DurationMinutes int       `gorm:"default:0" json:"durationMinutes"`
	IsActive        bool      `gorm:"default:true" json:"isActive"`
	SortOrder       int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	Category *ServiceCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (AddOnService) TableName() string { return "addon_services" }

// ServiceOption represents configurable options for a service
type ServiceOption struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ServiceID  uint      `gorm:"not null;index" json:"serviceId"`
	Name       string    `gorm:"type:varchar(255);not null" json:"name"`
	Type       string    `gorm:"type:varchar(50);not null" json:"type"` // select_single, select_multiple, quantity, text
	IsRequired bool      `gorm:"default:false" json:"isRequired"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	Service *Service              `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Choices []ServiceOptionChoice `gorm:"foreignKey:OptionID" json:"choices,omitempty"`
}

func (ServiceOption) TableName() string { return "service_options" }

// ServiceOptionChoice represents available choices for an option
type ServiceOptionChoice struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	OptionID                uint      `gorm:"not null;index" json:"optionId"`
	Label                   string    `gorm:"type:varchar(255);not null" json:"label"`
	PriceModifier           float64   `gorm:"type:decimal(10,2);default:0" json:"priceModifier"`
	DurationModifierMinutes int       `gorm:"default:0" json:"durationModifierMinutes"`
	CreatedAt               time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt               time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	Option *ServiceOption `gorm:"foreignKey:OptionID" json:"option,omitempty"`
}

func (ServiceOptionChoice) TableName() string { return "service_option_choices" }

// ServiceProvider represents a professional who can perform services
type ServiceProvider struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        string    `gorm:"type:uuid;uniqueIndex;not null" json:"userId"`
	Photo         *string   `gorm:"type:varchar(500)" json:"photo,omitempty"`
	Rating        float64   `gorm:"type:decimal(3,2);default:0" json:"rating"`
	Status        string    `gorm:"type:varchar(50);not null;default:'offline'" json:"status"` // available, busy, offline
	IsVerified    bool      `gorm:"default:false" json:"isVerified"`
	TotalJobs     int       `gorm:"default:0" json:"totalJobs"`
	CompletedJobs int       `gorm:"default:0" json:"completedJobs"`
	LastActive    time.Time `json:"lastActive"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (p *ServiceProvider) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
func (ServiceProvider) TableName() string { return "service_providers" }

// ServiceOrder represents a booking
type ServiceOrder struct {
	ID             string     `gorm:"type:uuid;primaryKey" json:"id"`
	Code           string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	UserID         string     `gorm:"type:uuid;not null;index" json:"userId"`
	ProviderID     *string    `gorm:"type:uuid;index" json:"providerId,omitempty"`
	Status         string     `gorm:"type:varchar(50);not null;index" json:"status"` // pending, searching, accepted, in_progress, completed, cancelled
	Address        string     `gorm:"type:text;not null" json:"address"`
	Latitude       float64    `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude      float64    `gorm:"type:decimal(11,8);not null" json:"longitude"`
	ServiceDate    time.Time  `gorm:"not null" json:"serviceDate"`
	Frequency      string     `gorm:"type:varchar(50);default:'once'" json:"frequency"`      // once, daily, weekly, monthly
	QuantityOfPros int        `gorm:"type:integer;not null;default:1" json:"quantityOfPros"` //  NEW: Number of professionals
	HoursOfService float64    `gorm:"type:decimal(5,2);not null;default:1.0" json:"hoursOfService"`
	CategorySlug   string     `gorm:"type:varchar(255);index" json:"categorySlug"` // Service category slug for provider filtering
	Subtotal       float64    `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	Discount       float64    `gorm:"type:decimal(10,2);default:0" json:"discount"`
	SurgeFee       float64    `gorm:"type:decimal(10,2);default:0" json:"surgeFee"`
	PlatformFee    float64    `gorm:"type:decimal(10,2);default:0" json:"platformFee"`
	Total          float64    `gorm:"type:decimal(10,2);not null" json:"total"`
	CouponCode     *string    `gorm:"type:varchar(50)" json:"couponCode,omitempty"`
	Notes          *string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	CancelledAt    *time.Time `json:"cancelledAt,omitempty"`
	WalletHold     float64    `gorm:"type:decimal(10,2);default:0" json:"walletHold"`
	WalletHoldID   *string    `gorm:"type:uuid" json:"walletHoldId,omitempty"`

	// Relations
	Items    []OrderItem      `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	AddOns   []OrderAddOn     `gorm:"foreignKey:OrderID" json:"addOns,omitempty"`
	Provider *ServiceProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}

// OrderAddOn - NEW: Track add-ons in orders
type OrderAddOn struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   string    `gorm:"type:uuid;not null;index" json:"orderId"`
	AddOnID   uint      `gorm:"not null" json:"addOnId"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	Order *ServiceOrder `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	AddOn *AddOnService `gorm:"foreignKey:AddOnID" json:"addOn,omitempty"`
}

func (OrderAddOn) TableName() string { return "order_add_ons" }

func (o *ServiceOrder) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	if o.Code == "" {
		o.Code = generateOrderCode()
	}
	return nil
}

func (ServiceOrder) TableName() string { return "service_orders" }

// OrderItem represents a service within an order
type OrderItem struct {
	ID              uint                   `gorm:"primaryKey" json:"id"`
	OrderID         string                 `gorm:"type:uuid;not null;index" json:"orderId"`
	ServiceID       string                 `gorm:"type:uuid;not null" json:"serviceId"`
	ServiceName     string                 `gorm:"type:varchar(255);not null" json:"serviceName"`
	BasePrice       float64                `gorm:"type:decimal(10,2);not null" json:"basePrice"`
	CalculatedPrice float64                `gorm:"type:decimal(10,2);not null" json:"calculatedPrice"`
	DurationMinutes int                    `gorm:"not null" json:"durationMinutes"`
	SelectedOptions map[string]interface{} `gorm:"type:jsonb" json:"selectedOptions"`
	CreatedAt       time.Time              `gorm:"autoCreateTime" json:"createdAt"`

	Order *ServiceOrder `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

func (OrderItem) TableName() string { return "order_items" }

// Rating represents a service rating
type Rating struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrderID    string    `gorm:"type:uuid;not null;uniqueIndex" json:"orderId"`
	UserID     string    `gorm:"type:uuid;not null;index" json:"userId"`
	ProviderID string    `gorm:"type:uuid;not null;index" json:"providerId"`
	Score      int       `gorm:"not null;check:score >= 1 AND score <= 5" json:"score"`
	Comment    *string   `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Rating) TableName() string { return "ratings" }

// SurgeZone for dynamic surge pricing (optional)
type SurgeZone struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"type:varchar(100);not null" json:"name"`
	SurgeMultiplier float64    `gorm:"type:decimal(3,2);default:1.00" json:"surgeMultiplier"`
	IsActive        bool       `gorm:"default:true" json:"isActive"`
	ValidFrom       *time.Time `json:"validFrom,omitempty"`
	ValidTo         *time.Time `json:"validTo,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (SurgeZone) TableName() string { return "surge_zones" }

func generateOrderCode() string {
	return fmt.Sprintf("HS%d", time.Now().Unix())
}
