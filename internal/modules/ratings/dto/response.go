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

type RatingStatsResponse struct {
	UserID           string  `json:"userId"`
	UserType         string  `json:"userType"` // 'driver' or 'rider'
	Rating           float64 `json:"rating"`
	TotalRides       int     `json:"totalRides"`
	TotalEarnings    float64 `json:"totalEarnings,omitempty"`
	TotalSpent       float64 `json:"totalSpent,omitempty"`
	CancellationRate float64 `json:"cancellationRate"`
	AcceptanceRate   float64 `json:"acceptanceRate,omitempty"`
}

type RatingBreakdownResponse struct {
	FiveStars     int     `json:"fiveStars"`
	FourStars     int     `json:"fourStars"`
	ThreeStars    int     `json:"threeStars"`
	TwoStars      int     `json:"twoStars"`
	OneStar       int     `json:"oneStar"`
	TotalRatings  int     `json:"totalRatings"`
	AverageRating float64 `json:"averageRating"`
}

func ToDriverRatingStats(profile *models.DriverProfile) *RatingStatsResponse {
	return &RatingStatsResponse{
		UserID:           profile.UserID,
		UserType:         "driver",
		Rating:           profile.Rating,
		TotalRides:       profile.TotalTrips,
		TotalEarnings:    profile.TotalEarnings,
		CancellationRate: profile.CancellationRate,
		AcceptanceRate:   profile.AcceptanceRate,
	}
}

func ToRiderRatingStats(profile *models.RiderProfile) *RatingStatsResponse {
	return &RatingStatsResponse{
		UserID:           profile.UserID,
		UserType:         "rider",
		Rating:           profile.Rating,
		TotalRides:       profile.TotalRides,
		TotalSpent:       profile.TotalSpent,
		CancellationRate: profile.CancellationRate,
	}
}
