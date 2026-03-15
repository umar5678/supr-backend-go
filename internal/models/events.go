package models

import (
	"time"

	"github.com/google/uuid"
)

type EventStatus string

const (
	EventStatusPending   EventStatus = "pending"
	EventStatusPublished EventStatus = "published"
	EventStatusDelivered EventStatus = "delivered"
	EventStatusFailed    EventStatus = "failed"
	EventStatusDLQ       EventStatus = "dlq"
)

type Event struct {
	ID           uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EventType    string      `gorm:"type:varchar(100);not null;index" json:"event_type"`
	SourceModule string      `gorm:"type:varchar(50);not null" json:"source_module"`
	Payload      []byte      `gorm:"type:jsonb;not null" json:"payload"`
	Status       EventStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	RetryCount   int         `gorm:"default:0" json:"retry_count"`
	FailedReason string      `gorm:"type:text" json:"failed_reason,omitempty"`
	CreatedAt    time.Time   `gorm:"autoCreateTime" json:"created_at"`
	PublishedAt  *time.Time  `json:"published_at,omitempty"`
}

func (Event) TableName() string {
	return "events"
}

type ProcessedEvent struct {
	EventID       uuid.UUID `gorm:"type:uuid;primary_key" json:"event_id"`
	ConsumerGroup string    `gorm:"type:varchar(100);primary_key" json:"consumer_group"`
	ProcessedAt   time.Time `gorm:"autoCreateTime" json:"processed_at"`
}

func (ProcessedEvent) TableName() string {
	return "processed_events"
}

type NotificationChannel string

const (
	ChannelInApp NotificationChannel = "in_app"
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelPush  NotificationChannel = "push"
)

type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusRead      NotificationStatus = "read"
)

type Notification struct {
	ID        uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	EventID   uuid.UUID           `gorm:"type:uuid;index" json:"event_id"`
	Channel   NotificationChannel `gorm:"type:varchar(20);not null" json:"channel"`
	Status    NotificationStatus  `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Title     string              `gorm:"type:varchar(255)" json:"title"`
	Message   string              `gorm:"type:text" json:"message"`
	Metadata  []byte              `gorm:"type:jsonb" json:"metadata,omitempty"`
	ReadAt    *time.Time          `json:"read_at,omitempty"`
	CreatedAt time.Time           `gorm:"autoCreateTime" json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}
