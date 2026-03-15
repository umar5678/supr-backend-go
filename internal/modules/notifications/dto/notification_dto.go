package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/umar5678/go-backend/internal/models"
)

type NotificationDTO struct {
	ID        uuid.UUID                  `json:"id"`
	EventID   uuid.UUID                  `json:"event_id"`
	Channel   models.NotificationChannel `json:"channel"`
	Status    models.NotificationStatus  `json:"status"`
	Title     string                     `json:"title"`
	Message   string                     `json:"message"`
	Metadata  map[string]interface{}     `json:"metadata,omitempty"`
	ReadAt    *time.Time                 `json:"read_at,omitempty"`
	CreatedAt time.Time                  `json:"created_at"`
}

type GetNotificationsRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

type GetNotificationsResponse struct {
	Notifications []*NotificationDTO `json:"notifications"`
	Total         int64              `json:"total"`
	Page          int                `json:"page"`
	PageSize      int                `json:"page_size"`
}

type UnreadCountResponse struct {
	Count int `json:"count"`
}

func ToNotificationDTO(n *models.Notification) *NotificationDTO {
	dto := &NotificationDTO{
		ID:        n.ID,
		EventID:   n.EventID,
		Channel:   n.Channel,
		Status:    n.Status,
		Title:     n.Title,
		Message:   n.Message,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}

	if len(n.Metadata) > 0 {
		// Parse metadata if needed
		dto.Metadata = make(map[string]interface{})
	}

	return dto
}
