// internal/modules/fraud/dto/response.go
package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

type FraudPatternResponse struct {
	ID            string     `json:"id"`
	PatternType   string     `json:"patternType"`
	UserID        *string    `json:"userId,omitempty"`
	DriverID      *string    `json:"driverId,omitempty"`
	RelatedUserID *string    `json:"relatedUserId,omitempty"`
	RideID        *string    `json:"rideId,omitempty"`
	Details       string     `json:"details"`
	RiskScore     int        `json:"riskScore"`
	Status        string     `json:"status"`
	ReviewedBy    *string    `json:"reviewedBy,omitempty"`
	ReviewNotes   string     `json:"reviewNotes,omitempty"`
	ReviewedAt    *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type FraudPatternListResponse struct {
	ID          string    `json:"id"`
	PatternType string    `json:"patternType"`
	UserID      *string   `json:"userId,omitempty"`
	RiskScore   int       `json:"riskScore"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type FraudStatsResponse struct {
	TotalPatterns      int            `json:"totalPatterns"`
	FlaggedCount       int            `json:"flaggedCount"`
	InvestigatingCount int            `json:"investigatingCount"`
	ConfirmedCount     int            `json:"confirmedCount"`
	DismissedCount     int            `json:"dismissedCount"`
	ByType             map[string]int `json:"byType"`
	HighRiskCount      int            `json:"highRiskCount"`
}

func ToFraudPatternResponse(pattern *models.FraudPattern) *FraudPatternResponse {
	return &FraudPatternResponse{
		ID:            pattern.ID,
		PatternType:   pattern.PatternType,
		UserID:        pattern.UserID,
		DriverID:      pattern.DriverID,
		RelatedUserID: pattern.RelatedUserID,
		RideID:        pattern.RideID,
		Details:       pattern.Details,
		RiskScore:     pattern.RiskScore,
		Status:        pattern.Status,
		ReviewedBy:    pattern.ReviewedBy,
		ReviewNotes:   pattern.ReviewNotes,
		ReviewedAt:    pattern.ReviewedAt,
		CreatedAt:     pattern.CreatedAt,
	}
}

func ToFraudPatternListResponse(pattern *models.FraudPattern) *FraudPatternListResponse {
	return &FraudPatternListResponse{
		ID:          pattern.ID,
		PatternType: pattern.PatternType,
		UserID:      pattern.UserID,
		RiskScore:   pattern.RiskScore,
		Status:      pattern.Status,
		CreatedAt:   pattern.CreatedAt,
	}
}
