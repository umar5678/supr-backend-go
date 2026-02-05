package batching

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/modules/batching/dto"
	"github.com/umar5678/go-backend/internal/modules/drivers"
	"github.com/umar5678/go-backend/internal/modules/ratings"
	"github.com/umar5678/go-backend/internal/modules/tracking"
	trackingdto "github.com/umar5678/go-backend/internal/modules/tracking/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type Service interface {
	AddRequestToBatch(ctx context.Context, req dto.RideRequestInfo) (batchID string, err error)
	GetBatchStatus(batchID string) (*dto.RequestBatch, error)
	GetBatchStatistics() *dto.BatchStatistics

	ProcessBatch(ctx context.Context, batchID string) (*dto.BatchMatchingResult, error)
	GetMatchingResult(batchID string) *dto.BatchMatchingResult

	SetBatchExpireCallback(callback func(batchID string))

	RankDriversForRequest(ctx context.Context, driverIDs []string, pickupLat, pickupLon float64) ([]dto.DriverRankingScore, error)
	GetDriverBreakdown(ctx context.Context, driverID string) (*dto.DriverRankingBreakdown, error)

	GetSystemStatus() map[string]interface{}
	Start() error
	Stop() error
}

type service struct {
	collector       *BatchCollector
	ranker          *DriverRanker
	matcher         *Matcher
	stats           *batchingStats
	trackingService tracking.Service
}

type batchingStats struct {
	TotalBatches        int64
	SuccessfulMatches   int64
	FailedMatches       int64
	AverageBatchSize    int64
	AverageMatchingTime int64
	LastUpdateTime      time.Time
}

func NewService(
	driversRepo drivers.Repository,
	ratingsService ratings.Service,
	trackingService tracking.Service,
	batchWindow time.Duration,
	maxBatchSize int,
) Service {
	collector := NewBatchCollector(batchWindow, maxBatchSize)

	ranker := NewDriverRanker(
		driversRepo,
		ratingsService,
		trackingService,
	)

	matcher := NewMatcher(ranker)

	return &service{
		collector:       collector,
		ranker:          ranker,
		matcher:         matcher,
		trackingService: trackingService,
		stats: &batchingStats{
			LastUpdateTime: time.Now(),
		},
	}
}

func (s *service) SetBatchExpireCallback(callback func(batchID string)) {
	s.collector.SetBatchExpireCallback(callback)
}

func (s *service) AddRequestToBatch(ctx context.Context, req dto.RideRequestInfo) (string, error) {
	return s.collector.AddRequest(ctx, req)
}

func (s *service) GetBatchStatus(batchID string) (*dto.RequestBatch, error) {
	return s.collector.GetBatch(batchID)
}

func (s *service) GetBatchStatistics() *dto.BatchStatistics {
	return &dto.BatchStatistics{
		TotalBatches:        s.stats.TotalBatches,
		AverageBatchSize:    float64(s.stats.AverageBatchSize),
		SuccessfulMatches:   s.stats.SuccessfulMatches,
		FailedMatches:       s.stats.FailedMatches,
		AverageMatchingTime: int(s.stats.AverageMatchingTime),
		LastUpdateTime:      s.stats.LastUpdateTime,
	}
}

func (s *service) ProcessBatch(ctx context.Context, batchID string) (*dto.BatchMatchingResult, error) {
	requests, err := s.collector.GetBatchRequests(batchID)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		result := &dto.BatchMatchingResult{
			BatchID:      batchID,
			Assignments:  []dto.OptimalAssignment{},
			UnmatchedIDs: []string{},
		}
		return result, nil
	}

	centroid := calculateCentroid(requests)

	radii := []float64{3.0, 5.0, 8.0}
	var nearbyDriverIDs []string

	for _, radiusKm := range radii {
		resp, err := s.trackingService.FindNearbyDrivers(ctx, trackingdto.FindNearbyDriversRequest{
			Latitude:      centroid.Latitude,
			Longitude:     centroid.Longitude,
			RadiusKm:      radiusKm,
			OnlyAvailable: true, 
			Limit:         50,   
		})
		if err == nil && resp != nil && len(resp.Drivers) > 0 {
			for _, driver := range resp.Drivers {
				nearbyDriverIDs = append(nearbyDriverIDs, driver.DriverID)
			}
			logger.Info("Found nearby drivers for batch",
				"batchID", batchID,
				"radiusKm", radiusKm,
				"driverCount", len(resp.Drivers),
			)
			break
		}
	}

	if len(nearbyDriverIDs) == 0 {
		logger.Warn("No drivers found for batch at any radius",
			"batchID", batchID,
			"requestCount", len(requests),
		)
		result := &dto.BatchMatchingResult{
			BatchID:     batchID,
			Assignments: []dto.OptimalAssignment{},
		}
		for _, req := range requests {
			result.UnmatchedIDs = append(result.UnmatchedIDs, req.RideID)
		}
		s.stats.FailedMatches += int64(len(requests))
		s.stats.LastUpdateTime = time.Now()
		return result, nil
	}

	rankedDrivers, err := s.ranker.RankDrivers(
		ctx,
		nearbyDriverIDs,
		centroid.Latitude,
		centroid.Longitude,
	)
	if err != nil {
		return nil, err
	}

	result := s.matcher.MatchRequestsToDrivers(
		ctx,
		requests,
		rankedDrivers,
		0.6,
	)

	s.stats.TotalBatches++
	s.stats.SuccessfulMatches += int64(len(result.Assignments))
	s.stats.FailedMatches += int64(len(result.UnmatchedIDs))
	if len(requests) > 0 {
		s.stats.AverageBatchSize = (s.stats.AverageBatchSize + int64(len(requests))) / 2
	}
	s.stats.LastUpdateTime = time.Now()

	s.collector.CompleteBatch(batchID)

	return result, nil
}

func (s *service) GetMatchingResult(batchID string) *dto.BatchMatchingResult {
	return nil
}

func (s *service) RankDriversForRequest(
	ctx context.Context,
	driverIDs []string,
	pickupLat, pickupLon float64,
) ([]dto.DriverRankingScore, error) {
	return s.ranker.RankDrivers(ctx, driverIDs, pickupLat, pickupLon)
}

func (s *service) GetDriverBreakdown(ctx context.Context, driverID string) (*dto.DriverRankingBreakdown, error) {
	return s.ranker.GetDetailedBreakdown(ctx, driverID)
}

func (s *service) GetSystemStatus() map[string]interface{} {
	return map[string]interface{}{
		"batches":               s.collector.GetBatchStatus(),
		"total_batches":         s.stats.TotalBatches,
		"successful_matches":    s.stats.SuccessfulMatches,
		"failed_matches":        s.stats.FailedMatches,
		"average_batch_size":    s.stats.AverageBatchSize,
		"average_match_time_ms": s.stats.AverageMatchingTime,
		"last_update":           s.stats.LastUpdateTime,
	}
}

func (s *service) Start() error {
	return nil
}

func (s *service) Stop() error {
	s.collector.Stop()
	return nil
}

func calculateCentroid(requests []dto.RideRequestInfo) dto.Location {
	if len(requests) == 0 {
		return dto.Location{}
	}

	var totalLat, totalLon float64
	for _, req := range requests {
		totalLat += req.PickupLat
		totalLon += req.PickupLon
	}

	return dto.Location{
		Latitude:  totalLat / float64(len(requests)),
		Longitude: totalLon / float64(len(requests)),
	}
}
