package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type MetaData map[string]interface{}

func (m MetaData) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MetaData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(bytes, &m)
}

type Document struct {
	ID                string                  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID            string                  `gorm:"type:uuid;not null;index" json:"userId"`
	User              User                    `gorm:"foreignKey:UserID" json:"-"`
	DriverID          *string                 `gorm:"type:uuid;index" json:"driverId,omitempty"`
	DriverProfile     *DriverProfile          `gorm:"foreignKey:DriverID" json:"-"`
	ServiceProviderID *string                 `gorm:"type:uuid;index" json:"serviceProviderId,omitempty"`
	ServiceProvider   *ServiceProviderProfile `gorm:"foreignKey:ServiceProviderID" json:"-"`

	DocumentType     string `gorm:"type:varchar(100);not null;index" json:"documentType"` 
	FileName         string `gorm:"type:varchar(255);not null" json:"fileName"`
	FileURL          string `gorm:"type:varchar(1000);not null" json:"fileUrl"`
	FileSize         int64  `gorm:"type:bigint" json:"fileSize"`
	MimeType         string `gorm:"type:varchar(50)" json:"mimeType"`
	ImageKitFileID   string `gorm:"type:varchar(255)" json:"-"`
	ImageKitFilePath string `gorm:"type:varchar(500)" json:"-"`

	// Verification status
	Status          string     `gorm:"type:varchar(50);default:'pending';index" json:"status"` 
	VerifiedBy      *string    `gorm:"type:uuid" json:"verifiedBy,omitempty"`
	VerifiedAdmin   *User      `gorm:"foreignKey:VerifiedBy" json:"-"`
	VerifiedAt      *time.Time `json:"verifiedAt,omitempty"`
	RejectionReason string     `gorm:"type:text" json:"rejectionReason,omitempty"`
	ExpiryDate      *time.Time `json:"expiryDate,omitempty"` 

	// Metadata
	IsFront  bool     `gorm:"default:false" json:"isFront,omitempty"`
	Metadata MetaData `gorm:"type:jsonb" json:"-"`                   

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Document) TableName() string {
	return "documents"
}

type DocumentVerificationLog struct {
	ID         string   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	DocumentID string   `gorm:"type:uuid;not null;index" json:"documentId"`
	Document   Document `gorm:"foreignKey:DocumentID" json:"-"`

	AdminID string `gorm:"type:uuid;not null" json:"adminId"`
	Admin   User   `gorm:"foreignKey:AdminID" json:"-"`

	Action   string `gorm:"type:varchar(50)" json:"action"` 
	Status   string `gorm:"type:varchar(50)" json:"status"` 
	Comments string `gorm:"type:text" json:"comments,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (DocumentVerificationLog) TableName() string {
	return "document_verification_logs"
}
