package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type MetaData map[string]interface{}

// Value implements the driver.Valuer interface for JSONB serialization
func (m MetaData) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB deserialization
func (m *MetaData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(bytes, &m)
}

// Document represents verification documents for service providers (drivers and service providers)
type Document struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID           string    `gorm:"type:uuid;not null;index" json:"userId"`
	User             User      `gorm:"foreignKey:UserID" json:"-"`
	DriverID         *string   `gorm:"type:uuid;index" json:"driverId,omitempty"`
	DriverProfile    *DriverProfile `gorm:"foreignKey:DriverID" json:"-"`
	ServiceProviderID *string  `gorm:"type:uuid;index" json:"serviceProviderId,omitempty"`
	ServiceProvider  *ServiceProviderProfile `gorm:"foreignKey:ServiceProviderID" json:"-"`
	
	// Document details
	DocumentType     string    `gorm:"type:varchar(100);not null;index" json:"documentType"` // license, aadhaar, registration, insurance, trade-license, profile-photo
	FileName         string    `gorm:"type:varchar(255);not null" json:"fileName"`
	FileURL          string    `gorm:"type:varchar(1000);not null" json:"fileUrl"`
	FileSize         int64     `gorm:"type:bigint" json:"fileSize"`
	MimeType         string    `gorm:"type:varchar(50)" json:"mimeType"`
	ImageKitFileID   string    `gorm:"type:varchar(255)" json:"-"` // ImageKit file ID for tracking
	ImageKitFilePath string    `gorm:"type:varchar(500)" json:"-"` // ImageKit file path (e.g., /documents/licenses/file.pdf)
	
	// Verification status
	Status           string    `gorm:"type:varchar(50);default:'pending';index" json:"status"` // pending, verified, rejected, expired
	VerifiedBy       *string   `gorm:"type:uuid" json:"verifiedBy,omitempty"`
	VerifiedAdmin    *User     `gorm:"foreignKey:VerifiedBy" json:"-"`
	VerifiedAt       *time.Time `json:"verifiedAt,omitempty"`
	RejectionReason  string    `gorm:"type:text" json:"rejectionReason,omitempty"`
	ExpiryDate       *time.Time `json:"expiryDate,omitempty"` // For documents with expiration dates
	
	// Metadata
	IsFront          bool      `gorm:"default:false" json:"isFront,omitempty"` // For dual documents (license front/back)
	Metadata         MetaData    `gorm:"type:jsonb" json:"-"` // Additional metadata as JSON
	
	// Timestamps
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Document) TableName() string {
	return "documents"
}

// DocumentVerificationLog tracks document verification history
type DocumentVerificationLog struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	DocumentID       string    `gorm:"type:uuid;not null;index" json:"documentId"`
	Document         Document  `gorm:"foreignKey:DocumentID" json:"-"`
	
	AdminID          string    `gorm:"type:uuid;not null" json:"adminId"`
	Admin            User      `gorm:"foreignKey:AdminID" json:"-"`
	
	Action           string    `gorm:"type:varchar(50)" json:"action"` // verified, rejected, re-requested
	Status           string    `gorm:"type:varchar(50)" json:"status"` // Status after action
	Comments         string    `gorm:"type:text" json:"comments,omitempty"`
	
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (DocumentVerificationLog) TableName() string {
	return "document_verification_logs"
}
