package dto

import (
	"time"
)

// RequestBatch represents a batch of ride requests collected within the batching window
type RequestBatch struct {
	ID              string    `json:"id"`
	RequestIDs      []string  `json:"request_ids"`
	PickupLocation  Location  `json:"pickup_location"`
	DropoffLocation Location  `json:"dropoff_location"`
	VehicleTypeID   string    `json:"vehicle_type_id"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	Status          string    `json:"status"` // pending, matching, assigned, expired
	RequestCount    int       `json:"request_count"`
}

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Geohash   string  `json:"geohash,omitempty"`
}

// RideRequestInfo contains basic ride request information for batching
type RideRequestInfo struct {
	RideID        string
	RiderID       string
	PickupLat     float64
	PickupLon     float64
	DropoffLat    float64
	DropoffLon    float64
	PickupGeohash string
	VehicleTypeID string
	RequestedAt   time.Time
}

// DriverRankingScore represents the comprehensive ranking score for a driver
type DriverRankingScore struct {
	DriverID          string  `json:"driver_id"`
	UserID            string  `json:"user_id"`
	RatingScore       float64 `json:"rating_score"`       // 0-40 (40% weight)
	AcceptanceScore   float64 `json:"acceptance_score"`   // 0-30 (30% weight)
	CancellationScore float64 `json:"cancellation_score"` // 0-20 (20% weight)
	CompletionScore   float64 `json:"completion_score"`   // 0-10 (10% weight)
	TotalScore        float64 `json:"total_score"`        // Sum of all above (0-100)
	Distance          float64 `json:"distance_km"`
	ResponseTime      int     `json:"response_time_seconds"`
	IsAvailable       bool    `json:"is_available"`
	RankPosition      int     `json:"rank_position,omitempty"`
}

// DriverRankingBreakdown contains detailed scoring information
type DriverRankingBreakdown struct {
	DriverID           string             `json:"driver_id"`
	UserID             string             `json:"user_id"`
	Rating             float64            `json:"rating"`
	RatingCount        int                `json:"rating_count"`
	AcceptanceRate     float64            `json:"acceptance_rate"`
	CancellationRate   float64            `json:"cancellation_rate"`
	CompletionRate     float64            `json:"completion_rate"`
	TotalRides         int                `json:"total_rides"`
	CompletedRides     int                `json:"completed_rides"`
	CancelledRides     int                `json:"cancelled_rides"`
	RejectedRequests   int                `json:"rejected_requests"`
	Distance           float64            `json:"distance_km"`
	EstimatedArrival   int                `json:"estimated_arrival_seconds"`
	IsOngoingRide      bool               `json:"is_ongoing_ride"`
	AvailabilityStatus string             `json:"availability_status"` // available, onride, offline
	Score              DriverRankingScore `json:"score"`
}

// GroupedRideRequests represents geographically grouped ride requests
type GroupedRideRequests struct {
	GroupID       string            `json:"group_id"`
	Requests      []RideRequestInfo `json:"requests"`
	Centroid      Location          `json:"centroid"`
	Radius        float64           `json:"radius_km"`
	VehicleTypeID string            `json:"vehicle_type_id"`
	RequestCount  int               `json:"request_count"`
}

// OptimalAssignment represents the best driver-request pair
type OptimalAssignment struct {
	RequestID  string  `json:"request_id"`
	RideID     string  `json:"ride_id"`
	DriverID   string  `json:"driver_id"`
	UserID     string  `json:"user_id"`
	Score      float64 `json:"score"`
	Distance   float64 `json:"distance_km"`
	ETA        int     `json:"eta_seconds"`
	Confidence float64 `json:"confidence"` // 0-1 how likely driver will accept
}

// BatchMatchingResult represents the result of batch matching
type BatchMatchingResult struct {
	BatchID      string              `json:"batch_id"`
	Assignments  []OptimalAssignment `json:"assignments"`
	UnmatchedIDs []string            `json:"unmatched_request_ids"`
	MatchedCount int                 `json:"matched_count"`
	Duration     int                 `json:"matching_duration_ms"`
}

// BatchStatistics tracks batching system metrics
type BatchStatistics struct {
	TotalBatches         int64     `json:"total_batches"`
	AverageBatchSize     float64   `json:"average_batch_size"`
	SuccessfulMatches    int64     `json:"successful_matches"`
	FailedMatches        int64     `json:"failed_matches"`
	AverageMatchingTime  int       `json:"average_matching_time_ms"`
	DriverAcceptanceRate float64   `json:"driver_acceptance_rate"`
	LastUpdateTime       time.Time `json:"last_update_time"`
}
