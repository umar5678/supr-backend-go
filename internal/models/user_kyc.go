package models

import (
    "time"
    "gorm.io/gorm"
)

type UserKYC struct {
    ID              string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    UserID          string         `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`
    User            User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
    IDType          string         `gorm:"type:varchar(50)" json:"idType"` // 'national_id', 'passport', 'driver_license'
    IDNumber        string         `gorm:"type:varchar(100)" json:"idNumber"`
    IDDocumentURL   string         `gorm:"type:varchar(500)" json:"idDocumentUrl"`
    SelfieURL       string         `gorm:"type:varchar(500)" json:"selfieUrl"`
    Status          string         `gorm:"type:varchar(50);default:'pending'" json:"status"` // 'pending', 'approved', 'rejected'
    RejectionReason string         `gorm:"type:text" json:"rejectionReason,omitempty"`
    VerifiedAt      *time.Time     `json:"verifiedAt,omitempty"`
    CreatedAt       time.Time      `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserKYC) TableName() string {
    return "user_kyc"
}
