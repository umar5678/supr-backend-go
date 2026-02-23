package models

import "time"

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeLocation MessageType = "location"
	MessageTypeStatus   MessageType = "status"
	MessageTypeSystem   MessageType = "system"
)

const (
	MessageEventNew      = "message:new"
	MessageEventRead     = "message:read"
	MessageEventDelete   = "message:delete"
	MessageEventTyping   = "message:typing"
	PresenceEventOnline  = "presence:online"
	PresenceEventOffline = "presence:offline"
)

type RideMessage struct {
	ID          string                 `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	RideID      string                 `json:"rideId" gorm:"type:uuid;index"`
	SenderID    string                 `json:"senderId" gorm:"type:uuid;index"`
	SenderType  string                 `json:"senderType"` 
	MessageType MessageType            `json:"messageType"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	IsRead      bool                   `json:"isRead" gorm:"default:false"`
	ReadAt      *time.Time             `json:"readAt"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	DeletedAt   *time.Time             `json:"-" gorm:"index"`

	Ride *Ride `json:"-" gorm:"foreignKey:RideID"`
}

type MessageResponse struct {
	ID          string                 `json:"id"`
	RideID      string                 `json:"rideId"`
	SenderID    string                 `json:"senderId"`
	SenderName  string                 `json:"senderName"`
	SenderType  string                 `json:"senderType"`
	MessageType string                 `json:"messageType"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IsRead      bool                   `json:"isRead"`
	ReadAt      *time.Time             `json:"readAt,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
}

type WSMessageEvent struct {
	Type     string      `json:"type"` 
	RideID   string      `json:"rideId"`
	SenderID string      `json:"senderId"`
	Data     interface{} `json:"data"`
}
