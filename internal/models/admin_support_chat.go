package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// AdminSupportChat represents a message in admin support chat conversation
type AdminSupportChat struct {
	ID              string  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ConversationID  string  `gorm:"type:uuid;not null;index" json:"conversationId"`
	ParentMessageID *string `gorm:"type:uuid;index" json:"parentMessageId,omitempty"`

	SenderID   string `gorm:"type:uuid;not null;index" json:"senderId"`
	SenderRole string `gorm:"type:varchar(50);not null" json:"senderRole"`
	Content    string `gorm:"type:text;not null" json:"content"`

	Metadata AdminSupportChatMetadata `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`

	IsRead        bool       `gorm:"default:false" json:"isRead"`
	ReadByAdminAt *time.Time `json:"readByAdminAt,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AdminSupportChat) TableName() string {
	return "admin_support_chats"
}

// AdminSupportChatMetadata holds additional metadata for a message
type AdminSupportChatMetadata struct {
	MessageID string                 `json:"messageId,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

func (m AdminSupportChatMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *AdminSupportChatMetadata) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(bytes, &m)
}
