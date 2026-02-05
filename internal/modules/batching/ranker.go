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

type DriverRanker struct {
	driversRepo     drivers.Repository
	ratingsService  ratings.Service
	trackingService tracking.Service

	ratingWeight       float64
	acceptanceWeight   float64
	cancellationWeight float64
	completionWeight   float64
}

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

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].TotalScore > scores[j].TotalScore
	})

	for i := range scores {
		scores[i].RankPosition = i + 1
	}

	return scores, nil
}

func (dr *DriverRanker) ScoreDriver(
	ctx context.Context,
	driverID string,
	pickupLat, pickupLon float64,
) (*dto.DriverRankingScore, error) {
	score := &dto.DriverRankingScore{
		DriverID: driverID,
	}

	driverProfile, err := dr.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	score.UserID = driverProfile.UserID

	locationResp, err := dr.trackingService.GetDriverLocation(ctx, driverID)
	if err != nil {
		return nil, err
	}

	distance := calculateHaversineDistance(
		pickupLat, pickupLon,
		locationResp.Latitude, locationResp.Longitude,
	)
	score.Distance = distance

	avgSpeedKmH := 40.0
	responseTime := int((distance / avgSpeedKmH) * 3600)
	if responseTime < 0 {
		responseTime = 0
	}
	score.ResponseTime = responseTime

	ratingScore, err := dr.calculateRatingScore(ctx, driverID)
	if err != nil {
		logger.Warn("failed to calculate rating score", "driverID", driverID, "error", err)
		ratingScore = 0
	}
	score.RatingScore = ratingScore

	acceptanceScore, err := dr.calculateAcceptanceScore(ctx, driverID)
	if err != nil {
		logger.Warn("failed to calculate acceptance score", "driverID", driverID, "error", err)
		acceptanceScore = 0
	}
	score.AcceptanceScore = acceptanceScore

	cancellationScore, err := dr.calculateCancellationScore(ctx, driverID)
	if err != nil {
		logger.Warn("failed to calculate cancellation score", "driverID", driverID, "error", err)
		cancellationScore = 0
	}
	score.CancellationScore = cancellationScore

	completionScore, err := dr.calculateCompletionScore(ctx, driverID)
	if err != nil {
		logger.Warn("failed to calculate completion score", "driverID", driverID, "error", err)
		completionScore = 0
	}
	score.CompletionScore = completionScore

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

func (dr *DriverRanker) calculateRatingScore(ctx context.Context, driverID string) (float64, error) {
	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return 0, err
	}

	if stats.Rating <= 0 {
		return 0, nil
	}

	score := (stats.Rating / 5.0) * dr.ratingWeight
	return score, nil
}

func (dr *DriverRanker) calculateAcceptanceScore(ctx context.Context, driverID string) (float64, error) {
	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return 0, err
	}

	acceptanceRate := stats.AcceptanceRate

	if acceptanceRate > 1.0 {
		acceptanceRate = 1.0
	}
	if acceptanceRate < 0 {
		acceptanceRate = 0
	}

	score := acceptanceRate * dr.acceptanceWeight
	return score, nil
}

func (dr *DriverRanker) calculateCancellationScore(ctx context.Context, driverID string) (float64, error) {
	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return 0, err
	}

	cancellationRate := stats.CancellationRate

	if cancellationRate > 1.0 {
		cancellationRate = 1.0
	}
	if cancellationRate < 0 {
		cancellationRate = 0
	}

	score := (1.0 - cancellationRate) * dr.cancellationWeight
	return score, nil
}

func (dr *DriverRanker) calculateCompletionScore(ctx context.Context, driverID string) (float64, error) {
	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return 0, err
	}

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

func (dr *DriverRanker) GetDetailedBreakdown(
	ctx context.Context,
	driverID string,
) (*dto.DriverRankingBreakdown, error) {
	breakdown := &dto.DriverRankingBreakdown{
		DriverID: driverID,
	}

	driverProfile, err := dr.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	breakdown.UserID = driverProfile.UserID

	stats, err := dr.ratingsService.GetDriverRatingStats(ctx, driverID)
	if err != nil {
		return nil, err
	}

	breakdown.Rating = stats.Rating
	breakdown.TotalRides = stats.TotalRides
	breakdown.AcceptanceRate = stats.AcceptanceRate
	breakdown.CancellationRate = stats.CancellationRate
	breakdown.CompletionRate = 1.0 - stats.CancellationRate

	score, err := dr.ScoreDriver(ctx, driverID, 0, 0)
	if err == nil {
		breakdown.Score = *score
	}

	return breakdown, nil
}

func calculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	a := math.Sin(dLat/2.0)*math.Sin(dLat/2.0) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2.0)*math.Sin(dLon/2.0)
	c := 2.0 * math.Asin(math.Sqrt(a))

	return earthRadiusKm * c
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}
