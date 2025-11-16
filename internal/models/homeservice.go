package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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

// ServiceCategory represents a category of home services
type ServiceCategory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IconURL     string    `gorm:"type:varchar(255)" json:"iconUrl"`
	IsActive    bool      `gorm:"default:true" json:"isActive"`
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ServiceCategory) TableName() string { return "service_categories" }

// Service represents a bookable home service (the "product")
type Service struct {
	ID                  uint             `gorm:"primaryKey" json:"id"`
	CategoryID          uint             `gorm:"not null" json:"categoryId"`
	Name                string           `gorm:"type:varchar(150);not null" json:"name"`
	Description         string           `gorm:"type:text" json:"description"`
	ImageURL            string           `gorm:"type:varchar(255)" json:"imageUrl"`
	BasePrice           float64          `gorm:"type:decimal(10,2);not null" json:"basePrice"`
	PricingModel        string           `gorm:"type:varchar(50);not null" json:"pricingModel"` // 'fixed', 'hourly', 'per_unit'
	BaseDurationMinutes int              `gorm:"not null" json:"baseDurationMinutes"`
	IsActive            bool             `gorm:"default:true" json:"isActive"`
	CreatedAt           time.Time        `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time        `gorm:"autoUpdateTime" json:"updatedAt"`
	Category            *ServiceCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Options             []ServiceOption  `gorm:"foreignKey:ServiceID" json:"options,omitempty"`
}

func (Service) TableName() string { return "services" }

// ServiceOption represents configurable options for a service
type ServiceOption struct {
	ID         uint                  `gorm:"primaryKey" json:"id"`
	ServiceID  uint                  `gorm:"not null" json:"serviceId"`
	Name       string                `gorm:"type:varchar(100);not null" json:"name"`
	Type       string                `gorm:"type:varchar(50);not null" json:"type"` // 'select_single', 'select_multiple', 'quantity', 'text'
	IsRequired bool                  `gorm:"default:false" json:"isRequired"`
	CreatedAt  time.Time             `gorm:"autoCreateTime" json:"createdAt"`
	Choices    []ServiceOptionChoice `gorm:"foreignKey:OptionID" json:"choices,omitempty"`
}

func (ServiceOption) TableName() string { return "service_options" }

// ServiceOptionChoice represents available choices for an option
type ServiceOptionChoice struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	OptionID                uint      `gorm:"not null" json:"optionId"`
	Label                   string    `gorm:"type:varchar(100);not null" json:"label"`
	PriceModifier           float64   `gorm:"type:decimal(10,2);default:0.00" json:"priceModifier"`
	DurationModifierMinutes int       `gorm:"default:0" json:"durationModifierMinutes"`
	CreatedAt               time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (ServiceOptionChoice) TableName() string { return "service_option_choices" }

// ServiceProvider represents a professional who can perform services
type ServiceProvider struct {
	ID            string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID        string     `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`
	Photo         *string    `gorm:"type:varchar(255)" json:"photo,omitempty"`
	Rating        float64    `gorm:"type:decimal(3,2);default:5.00" json:"rating"`
	Status        string     `gorm:"type:varchar(50);default:'offline'" json:"status"` // 'available', 'busy', 'offline'
	IsVerified    bool       `gorm:"default:false" json:"isVerified"`
	TotalJobs     int        `gorm:"default:0" json:"totalJobs"`
	CompletedJobs int        `gorm:"default:0" json:"completedJobs"`
	LastActive    *time.Time `json:"lastActive,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	User          *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	// Location stored separately in PostGIS
	QualifiedServices []Service `gorm:"many2many:provider_qualified_services;foreignKey:ID;joinForeignKey:ProviderID;References:ID;joinReferences:ServiceID" json:"qualifiedServices,omitempty"`
}

func (ServiceProvider) TableName() string { return "service_providers" }

// ServiceOrder represents a booking
type ServiceOrder struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Code         string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"`
	UserID       string         `gorm:"type:uuid;not null;index" json:"userId"`
	ProviderID   *string        `gorm:"type:uuid;index" json:"providerId,omitempty"`
	Status       string         `gorm:"type:varchar(50);not null;index" json:"status"` // 'pending', 'searching_provider', 'accepted', 'in_progress', 'completed', 'cancelled', 'no_provider_available'
	Address      string         `gorm:"type:text;not null" json:"address"`
	ServiceDate  time.Time      `gorm:"not null" json:"serviceDate"`
	Notes        *string        `gorm:"type:text" json:"notes,omitempty"`
	Frequency    string         `gorm:"type:varchar(20);default:'once'" json:"frequency"` // 'once', 'daily', 'weekly', 'monthly'
	Subtotal     float64        `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	Discount     float64        `gorm:"type:decimal(10,2);default:0.00" json:"discount"`
	SurgeFee     float64        `gorm:"type:decimal(10,2);default:0.00" json:"surgeFee"`
	PlatformFee  float64        `gorm:"type:decimal(10,2);default:0.00" json:"platformFee"`
	Total        float64        `gorm:"type:decimal(10,2);not null" json:"total"`
	CouponCode   *string        `gorm:"type:varchar(50)" json:"couponCode,omitempty"`
	WalletHoldID *string        `gorm:"type:uuid" json:"walletHoldId,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;index" json:"createdAt"`
	AcceptedAt   *time.Time     `json:"acceptedAt,omitempty"`
	StartedAt    *time.Time     `json:"startedAt,omitempty"`
	CompletedAt  *time.Time     `json:"completedAt,omitempty"`
	CancelledAt  *time.Time     `json:"cancelledAt,omitempty"`
	Items        []OrderItem    `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	User         *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Provider     *ServiceProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	WalletHold   *WalletHold    `gorm:"foreignKey:WalletHoldID" json:"walletHold,omitempty"`
}

func (ServiceOrder) TableName() string { return "service_orders" }

// OrderItem represents a service within an order
type OrderItem struct {
	ID              uint     `gorm:"primaryKey" json:"id"`
	OrderID         string   `gorm:"type:uuid;not null;index" json:"orderId"`
	ServiceID       uint     `gorm:"not null" json:"serviceId"`
	ServiceName     string   `gorm:"type:varchar(150);not null" json:"serviceName"` // Denormalized
	BasePrice       float64  `gorm:"type:decimal(10,2);not null" json:"basePrice"`
	CalculatedPrice float64  `gorm:"type:decimal(10,2);not null" json:"calculatedPrice"`
	DurationMinutes int      `gorm:"not null" json:"durationMinutes"`
	SelectedOptions JSONBMap `gorm:"type:jsonb" json:"selectedOptions"`
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