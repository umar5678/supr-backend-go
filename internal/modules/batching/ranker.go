package batching

import (
	"context"
	"math"
	"sort"

	"github.com/umar5678/go-backend/internal/modules/batching/dto"
	"github.com/umar5678/go-backend/internal/modules/drivers"
	"github.com/umar5678/go-backend/internal/modules/ratings"
	"github.com/umar5678/go-backend/internal/modules/tracking"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// DriverRanker implements the 4-factor driver ranking algorithm
type DriverRanker struct {
	driversRepo     drivers.Repository
	ratingsService  ratings.Service
	trackingService tracking.Service

	// Weighting factors (sum = 100)
	ratingWeight       float64 // 40%
	acceptanceWeight   float64 // 30%
	cancellationWeight float64 // 20%
	completionWeight   float64 // 10%
}

// NewDriverRanker creates a new driver ranking engine
func NewDriverRanker(
	driversRepo drivers.Repository,
	ratingsService ratings.Service,
	trackingService tracking.Service,
) *DriverRanker {
	return &DriverRanker{
		driversRepo:        driversRepo,
		ratingsService:     ratingsService,
		trackingService:    trackingService,
		ratingWeight:       40.0,
		acceptanceWeight:   30.0,
		cancellationWeight: 20.0,
		completionWeight:   10.0,
	}
}

// RankDrivers ranks a list of nearby drivers based on the 4-factor algorithm
func (dr *DriverRanker) RankDrivers(
	ctx context.Context,
	driverIDs []string,
	pickupLat, pickupLon float64,
) ([]dto.DriverRankingScore, error) {
	if len(driverIDs) == 0 {
		return []dto.DriverRankingScore{}, nil
	}

	scores := make([]dto.DriverRankingScore, 0, len(driverIDs))

	for _, driverID := range driverIDs {
		score, err := dr.ScoreDriver(ctx, driverID, pickupLat, pickupLon)
		if err != nil {
			logger.Warn("failed to score driver",
				"driverID", driverID,
				"error", err,
			)
			continue
		}

		scores = append(scores, *score)
	}

	// Sort by total score descending (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].TotalScore > scores[j].TotalScore
	})

	// Set rank positions
	for i := range scores {
		scores[i].RankPosition = i + 1
	}

	return scores, nil
}

// ScoreDriver calculates the comprehensive ranking score for a single driver
func (dr *DriverRanker) ScoreDriver(
	ctx context.Context,
	driverID string,
	pickupLat, pickupLon float64,
) (*dto.DriverRankingScore, error) {
	score := &dto.DriverRankingScore{
		DriverID: driverID,
	}

	// Get driver profile for user ID
	driverProfile, err := dr.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	score.UserID = driverProfile.UserID

	// Get driver location for distance
	locationResp, err := dr.trackingService.GetDriverLocation(ctx, driverID)
	if err != nil {
		return nil, err
	}

	// Calculate distance
	distance := calculateHaversineDistance(
		pickupLat, pickupLon,
		locationResp.Latitude, locationResp.Longitude,
	)
	score.Distance = distance

	// Calculate response time (ETA in seconds)
	// Assuming average speed of 40 km/h in urban areas
	avgSpeedKmH := 40.0
	responseTime := int((distance / avgSpeedKmH) * 3600) // Convert to seconds
	if responseTime < 0 {
		responseTime = 0
	}
	score.ResponseTime = responseTime

	// 1. RATING SCORE (40% weight, max 40 points)
	ratingScore, err := dr.calculateRatingScore(ctx, driverID)
	if err != nil {
		logger.Warn("failed to calculate rating score", "driverID", driverID, "error", err)
		ratingScore = 0
	}
	score.RatingScore = ratingScore

	// 2. ACCEPTANCE SCORE (30% weight, max 30 points)
	acceptanceScore, err := dr.calculateAcceptanceScore(ctx, driverID)
	if err != nil {
		logger.Warn("failed to calculate acceptance score", "driverID", driverID, "error", err)
		acceptanceScore = 0
	}
	score.AcceptanceScore = acceptanceScore

	// 3. CANCELLATION SCORE (20% weight, max 20 points)
	// Lower cancellation rate = higher score
	cancellationScore, err := dr.calculateCancellationScore(ctx, driverID)
	if err != nil {
		logger.Warn("failed to calculate cancellation score", "driverID", driverID, "error", err)
		cancellationScore = 0
	}
	score.CancellationScore = cancellationScore

	// 4. COMPLETION SCORE (10% weight, max 10 points)
	completionScore, err := dr.calculateCompletionScore(ctx, driverID)
	if err != nil {
		logger.Warn("failed to calculate completion score", "driverID", driverID, "error", err)
		completionScore = 0
	}
	score.CompletionScore = completionScore

	// Total score (sum of weighted scores)
	score.TotalScore = score.RatingScore + score.AcceptanceScore +
		score.CancellationScore + score.CompletionScore

	logger.Debug("Driver scored",
		"driverID", driverID,
		"totalScore", score.TotalScore,
		"ratingScore", score.RatingScore,
		"acceptanceScore", score.AcceptanceScore,
		"cancellationScore", score.CancellationScore,
		"completionScore", score.CompletionScore,
		"distance", distance,
		"responseTime", responseTime,
	)

	return score, nil
}

// calculateRatingScore calculates the 40-point rating component
// Based on 5-star system: rating/5 * 40
func (dr *DriverRanker) calculateRatingScore(ctx context.Context, driverID string) (float64, error) {
	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return 0, err
	}

	// Convert 5-star rating to 40-point scale
	// 5.0 stars = 40 points
	// 1.0 star = 8 points
	// 0 stars = 0 points
	if stats.Rating <= 0 {
		return 0, nil
	}

	score := (stats.Rating / 5.0) * dr.ratingWeight
	return score, nil
}

// calculateAcceptanceScore calculates the 30-point acceptance component
// Based on acceptance rate: (acceptanceRate) * 30
// Acceptance rate = accepted requests / total requests sent
func (dr *DriverRanker) calculateAcceptanceScore(ctx context.Context, driverID string) (float64, error) {
	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return 0, err
	}

	// Use acceptance rate from stats
	acceptanceRate := stats.AcceptanceRate

	// Cap at 1.0 and convert to 30-point scale
	if acceptanceRate > 1.0 {
		acceptanceRate = 1.0
	}
	if acceptanceRate < 0 {
		acceptanceRate = 0
	}

	score := acceptanceRate * dr.acceptanceWeight
	return score, nil
}

// calculateCancellationScore calculates the 20-point cancellation component
// Based on cancellation rate: (1 - cancellationRate) * 20
// Cancellation rate = cancelled rides / total rides
// Lower cancellation = higher score
func (dr *DriverRanker) calculateCancellationScore(ctx context.Context, driverID string) (float64, error) {
	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return 0, err
	}

	// Use cancellation rate from stats
	cancellationRate := stats.CancellationRate

	// Cap values
	if cancellationRate > 1.0 {
		cancellationRate = 1.0
	}
	if cancellationRate < 0 {
		cancellationRate = 0
	}

	// Inverse: lower cancellation = higher score
	score := (1.0 - cancellationRate) * dr.cancellationWeight
	return score, nil
}

// calculateCompletionScore calculates the 10-point completion component
// Based on completion rate: (completedRides / totalRides) * 10
func (dr *DriverRanker) calculateCompletionScore(ctx context.Context, driverID string) (float64, error) {
	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return 0, err
	}

	// Calculate completion rate from total rides and cancellation rate
	// Completed = Total - (Total * CancellationRate)
	var completionRate float64
	if stats.TotalRides > 0 {
		completionRate = 1.0 - stats.CancellationRate
	}

	// Cap at 1.0
	if completionRate > 1.0 {
		completionRate = 1.0
	}
	if completionRate < 0 {
		completionRate = 0
	}

	score := completionRate * dr.completionWeight
	return score, nil
}

// GetDetailedBreakdown returns a detailed breakdown of driver's scoring metrics
func (dr *DriverRanker) GetDetailedBreakdown(
	ctx context.Context,
	driverID string,
) (*dto.DriverRankingBreakdown, error) {
	breakdown := &dto.DriverRankingBreakdown{
		DriverID: driverID,
	}

	// Get driver profile for user ID
	driverProfile, err := dr.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	breakdown.UserID = driverProfile.UserID

	// Get rating stats
	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return nil, err
	}

	breakdown.Rating = stats.Rating
	breakdown.TotalRides = stats.TotalRides
	breakdown.AcceptanceRate = stats.AcceptanceRate
	breakdown.CancellationRate = stats.CancellationRate
	breakdown.CompletionRate = 1.0 - stats.CancellationRate

	// Calculate score
	score, err := dr.ScoreDriver(ctx, driverID, 0, 0) // Lat/Lon will be updated contextually
	if err == nil {
		breakdown.Score = *score
	}

	return breakdown, nil
}

// calculateHaversineDistance calculates distance between two geographic points (in km)
func calculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	// Convert degrees to radians
	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	// Haversine formula
	a := math.Sin(dLat/2.0)*math.Sin(dLat/2.0) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2.0)*math.Sin(dLon/2.0)
	c := 2.0 * math.Asin(math.Sqrt(a))

	return earthRadiusKm * c
}

// degreesToRadians converts degrees to radians
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}
