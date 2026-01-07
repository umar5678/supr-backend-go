package models

import (
    "time"
    "gorm.io/gorm"
)

type SavedLocation struct {
    ID         string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    UserID     string         `gorm:"type:uuid;not null;index" json:"userId"`
    User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Label      string         `gorm:"type:varchar(50);not null" json:"label"` // 'home', 'work', 'custom'
    CustomName string         `gorm:"type:varchar(100)" json:"customName,omitempty"`
    Address    string         `gorm:"type:varchar(500);not null" json:"address"`
    Latitude   float64        `gorm:"type:decimal(10,8);not null" json:"latitude"`
    Longitude  float64        `gorm:"type:decimal(11,8);not null" json:"longitude"`
    IsDefault  bool           `gorm:"default:false" json:"isDefault"`
    CreatedAt  time.Time      `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
    DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SavedLocation) TableName() string {
    return "saved_locations"
}
