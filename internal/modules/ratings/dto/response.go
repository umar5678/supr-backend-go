package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

type RatingResponse struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"orderId"`
	UserID     string    `json:"userId"`
	ProviderID string    `json:"providerId"`
	Score      int       `json:"score"`
	Comment    *string   `json:"comment,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

func ToRatingResponse(rating *models.Rating) *RatingResponse {
	return &RatingResponse{
		ID:         rating.ID,
		OrderID:    rating.OrderID,
		UserID:     rating.UserID,
		ProviderID: rating.ProviderID,
		Score:      rating.Score,
		Comment:    rating.Comment,
		CreatedAt:  rating.CreatedAt,
	}
}
