package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Todo struct {
	ID          string         `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Completed   bool           `gorm:"default:false" json:"completed"`
	UserID      string         `gorm:"type:uuid;not null" json:"user_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *Todo) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

func (Todo) TableName() string {
	return "todos"
}
