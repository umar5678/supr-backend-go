package dto

import "time"

type MessageResponse struct {
	ID        string                 `json:"id"`
	RideID    string                 `json:"rideId"`
	SenderID  string                 `json:"senderId"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}
