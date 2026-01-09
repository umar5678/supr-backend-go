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

// Service is the main batching service interface
type Service interface {
	// Batch collection
	AddRequestToBatch(ctx context.Context, req dto.RideRequestInfo) (batchID string, err error)
	GetBatchStatus(batchID string) (*dto.RequestBatch, error)
	GetBatchStatistics() *dto.BatchStatistics

	// Batch processing
	ProcessBatch(ctx context.Context, batchID string) (*dto.BatchMatchingResult, error)
	GetMatchingResult(batchID string) *dto.BatchMatchingResult

	// Batch callbacks
	SetBatchExpireCallback(callback func(batchID string)) // Called when batch expires

	// Driver ranking and matching
	RankDriversForRequest(ctx context.Context, driverIDs []string, pickupLat, pickupLon float64) ([]dto.DriverRankingScore, error)
	GetDriverBreakdown(ctx context.Context, driverID string) (*dto.DriverRankingBreakdown, error)

	// Status and health
	GetSystemStatus() map[string]interface{}
	Start() error
	Stop() error
}

type service struct {
	collector       *BatchCollector
	ranker          *DriverRanker
	matcher         *Matcher
	stats           *batchingStats
	trackingService tracking.Service // For finding nearby drivers
}

// batchingStats tracks system-wide metrics
type batchingStats struct {
	TotalBatches        int64
	SuccessfulMatches   int64
	FailedMatches       int64
	AverageBatchSize    int64
	AverageMatchingTime int64
	LastUpdateTime      time.Time
}

// NewService creates a new batching service
func NewService(
	driversRepo drivers.Repository,
	ratingsService ratings.Service,
	trackingService tracking.Service,
	batchWindow time.Duration,
	maxBatchSize int,
) Service {
	collector := NewBatchCollector(batchWindow, maxBatchSize)

	// Create ranker with properly typed parameters
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

// SetBatchExpireCallback sets the callback function to call when a batch expires
func (s *service) SetBatchExpireCallback(callback func(batchID string)) {
	s.collector.SetBatchExpireCallback(callback)
}

// AddRequestToBatch adds a ride request to the current batch for its vehicle type
func (s *service) AddRequestToBatch(ctx context.Context, req dto.RideRequestInfo) (string, error) {
	return s.collector.AddRequest(ctx, req)
}

// GetBatchStatus returns the current status of a batch
func (s *service) GetBatchStatus(batchID string) (*dto.RequestBatch, error) {
	return s.collector.GetBatch(batchID)
}

// GetBatchStatistics returns current system statistics
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

// ProcessBatch processes a batch for matching
// This should be called when batch window expires or max size is reached
func (s *service) ProcessBatch(ctx context.Context, batchID string) (*dto.BatchMatchingResult, error) {
	// Get batch requests
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

	// Get pickup centroid for driver search
	centroid := calculateCentroid(requests)

	// ✅ FIXED: Find nearby drivers using tracking service
	// Search at expanding radii: 3km, 5km, 8km
	radii := []float64{3.0, 5.0, 8.0} // kilometers
	var nearbyDriverIDs []string

	for _, radiusKm := range radii {
		resp, err := s.trackingService.FindNearbyDrivers(ctx, trackingdto.FindNearbyDriversRequest{
			Latitude:      centroid.Latitude,
			Longitude:     centroid.Longitude,
			RadiusKm:      radiusKm,
			OnlyAvailable: true, // Only get drivers not on active ride
			Limit:         50,   // Get up to 50 nearby drivers
		})
		if err == nil && resp != nil && len(resp.Drivers) > 0 {
			// Collect driver IDs from tracking response
			for _, driver := range resp.Drivers {
				nearbyDriverIDs = append(nearbyDriverIDs, driver.DriverID)
			}
			logger.Info("Found nearby drivers for batch",
				"batchID", batchID,
				"radiusKm", radiusKm,
				"driverCount", len(resp.Drivers),
			)
			// Found some drivers, proceed with matching
			break
		}
	}

	if len(nearbyDriverIDs) == 0 {
		// No drivers found at any radius
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

	// Rank drivers for this batch
	rankedDrivers, err := s.ranker.RankDrivers(
		ctx,
		nearbyDriverIDs,
		centroid.Latitude,
		centroid.Longitude,
	)
	if err != nil {
		return nil, err
	}

	// Match requests to drivers
	result := s.matcher.MatchRequestsToDrivers(
		ctx,
		requests,
		rankedDrivers,
		0.6, // 60% confidence threshold
	)

	// Update statistics
	s.stats.TotalBatches++
	s.stats.SuccessfulMatches += int64(len(result.Assignments))
	s.stats.FailedMatches += int64(len(result.UnmatchedIDs))
	if len(requests) > 0 {
		s.stats.AverageBatchSize = (s.stats.AverageBatchSize + int64(len(requests))) / 2
	}
	s.stats.LastUpdateTime = time.Now()

	// Complete batch
	s.collector.CompleteBatch(batchID)

	return result, nil
}

// GetMatchingResult returns the matching result for a batch (stub for now)
func (s *service) GetMatchingResult(batchID string) *dto.BatchMatchingResult {
	// TODO: Persist and retrieve matching results
	return nil
}

// RankDriversForRequest ranks drivers for a specific request
func (s *service) RankDriversForRequest(
	ctx context.Context,
	driverIDs []string,
	pickupLat, pickupLon float64,
) ([]dto.DriverRankingScore, error) {
	return s.ranker.RankDrivers(ctx, driverIDs, pickupLat, pickupLon)
}

// GetDriverBreakdown returns detailed breakdown of driver's scores
func (s *service) GetDriverBreakdown(ctx context.Context, driverID string) (*dto.DriverRankingBreakdown, error) {
	return s.ranker.GetDetailedBreakdown(ctx, driverID)
}

// GetSystemStatus returns overall system status
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

// Start initializes the batching service
func (s *service) Start() error {
	// Start background processes if needed
	return nil
}

// Stop gracefully shuts down the batching service
func (s *service) Stop() error {
	s.collector.Stop()
	return nil
}

// calculateCentroid calculates the geographic center of requests
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
