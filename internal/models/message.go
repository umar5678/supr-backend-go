package models

import "time"

type MessageType string

const (
    MessageTypeText     MessageType = "text"
    MessageTypeLocation MessageType = "location"
    MessageTypeStatus   MessageType = "status"
    MessageTypeSystem   MessageType = "system"
)

// WebSocket message event types
const (
    MessageEventNew    = "message:new"
    MessageEventRead   = "message:read"
    MessageEventDelete = "message:delete"
    MessageEventTyping = "message:typing"
    PresenceEventOnline  = "presence:online"
    PresenceEventOffline = "presence:offline"
)

type RideMessage struct {
    ID        string        `json:"id" gorm:"primaryKey"`
    RideID    string        `json:"rideId" gorm:"index"`
    SenderID  string        `json:"senderId" gorm:"index"`
    SenderType string      `json:"senderType"` // "rider" or "driver"
    MessageType MessageType `json:"messageType"`
    Content   string        `json:"content"`
    Metadata  map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    IsRead    bool          `json:"isRead" gorm:"default:false"`
    ReadAt    *time.Time    `json:"readAt"`
    CreatedAt time.Time     `json:"createdAt"`
    UpdatedAt time.Time     `json:"updatedAt"`
    DeletedAt *time.Time    `json:"-" gorm:"index"`

    // Relations
    Ride *Ride `json:"-" gorm:"foreignKey:RideID"`
}

// MessageResponse DTO
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

// WebSocket message event
type WSMessageEvent struct {
    Type    string      `json:"type"` // "message", "typing", "read", "status"
    RideID  string      `json:"rideId"`
    SenderID string     `json:"senderId"`
    Data    interface{} `json:"data"`
}