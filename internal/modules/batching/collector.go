package batching

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/modules/batching/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// BatchCollector manages the collection of ride requests within a batching window
type BatchCollector struct {
	batchWindow        time.Duration     // 10-second default window
	maxBatchSize       int               // maximum requests per batch
	batches            map[string]*batch // batchID -> batch
	vehicleTypeBatches map[string]string // vehicleTypeID -> current batchID
	mu                 sync.RWMutex
	ticker             *time.Ticker
	done               chan struct{}
	onBatchExpire      func(batchID string) // Callback when batch expires for processing
}

type batch struct {
	ID            string
	Requests      []dto.RideRequestInfo
	CreatedAt     time.Time
	ExpiresAt     time.Time
	VehicleTypeID string
	mu            sync.Mutex
}

// NewBatchCollector creates a new batch collector with 10-second window
// onBatchExpire is a callback function that's called when a batch window expires
// The callback should call ProcessBatch() to match requests to drivers
func NewBatchCollector(batchWindow time.Duration, maxBatchSize int) *BatchCollector {
	bc := &BatchCollector{
		batchWindow:        batchWindow,
		maxBatchSize:       maxBatchSize,
		batches:            make(map[string]*batch),
		vehicleTypeBatches: make(map[string]string),
		done:               make(chan struct{}),
		onBatchExpire:      nil, // Set later via SetBatchExpireCallback
	}

	// Start cleanup goroutine
	go bc.cleanupExpiredBatches()

	return bc
}

// SetBatchExpireCallback sets the callback function to call when a batch expires
func (bc *BatchCollector) SetBatchExpireCallback(callback func(batchID string)) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.onBatchExpire = callback
}

// AddRequest adds a ride request to the appropriate batch
func (bc *BatchCollector) AddRequest(ctx context.Context, req dto.RideRequestInfo) (string, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Get or create batch for this vehicle type
	batchID, exists := bc.vehicleTypeBatches[req.VehicleTypeID]
	var b *batch

	if !exists {
		// Create new batch
		batchID = uuid.New().String()
		b = &batch{
			ID:            batchID,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(bc.batchWindow),
			VehicleTypeID: req.VehicleTypeID,
			Requests:      []dto.RideRequestInfo{},
		}
		bc.batches[batchID] = b
		bc.vehicleTypeBatches[req.VehicleTypeID] = batchID

		logger.Info("Created new batch",
			"batchID", batchID,
			"vehicleTypeID", req.VehicleTypeID,
			"windowSeconds", int(bc.batchWindow.Seconds()),
		)
	} else {
		b = bc.batches[batchID]
	}

	// Add request to batch
	b.mu.Lock()
	b.Requests = append(b.Requests, req)
	requestCount := len(b.Requests)
	b.mu.Unlock()

	logger.Debug("Added request to batch",
		"batchID", batchID,
		"rideID", req.RideID,
		"requestCount", requestCount,
	)

	// Return batch info immediately but don't block on timing
	return batchID, nil
}

// GetBatch retrieves a batch by ID
func (bc *BatchCollector) GetBatch(batchID string) (*dto.RequestBatch, error) {
	bc.mu.RLock()
	b, exists := bc.batches[batchID]
	bc.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Calculate centroid
	centroid := bc.calculateCentroid(b.Requests)

	return &dto.RequestBatch{
		ID:             b.ID,
		RequestIDs:     extractRequestIDs(b.Requests),
		PickupLocation: centroid,
		VehicleTypeID:  b.VehicleTypeID,
		CreatedAt:      b.CreatedAt,
		ExpiresAt:      b.ExpiresAt,
		RequestCount:   len(b.Requests),
		Status:         "pending",
	}, nil
}

// GetBatchRequests returns all requests in a batch
func (bc *BatchCollector) GetBatchRequests(batchID string) ([]dto.RideRequestInfo, error) {
	bc.mu.RLock()
	b, exists := bc.batches[batchID]
	bc.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Return copy to prevent external modifications
	requests := make([]dto.RideRequestInfo, len(b.Requests))
	copy(requests, b.Requests)

	return requests, nil
}

// CheckBatchReady checks if batch is ready for processing (window expired or max size reached)
func (bc *BatchCollector) CheckBatchReady(batchID string) bool {
	bc.mu.RLock()
	b, exists := bc.batches[batchID]
	bc.mu.RUnlock()

	if !exists {
		return false
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Ready if window expired or max size reached
	isExpired := time.Now().After(b.ExpiresAt)
	isMaxed := len(b.Requests) >= bc.maxBatchSize

	return isExpired || isMaxed
}

// CompleteBatch marks batch as processed and removes it
func (bc *BatchCollector) CompleteBatch(batchID string) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if b, exists := bc.batches[batchID]; exists {
		// Remove from vehicle type index
		delete(bc.vehicleTypeBatches, b.VehicleTypeID)
		// Remove batch
		delete(bc.batches, batchID)

		logger.Info("Completed batch",
			"batchID", batchID,
			"vehicleTypeID", b.VehicleTypeID,
		)
	}

	return nil
}

// GetBatchStatus returns current status of all batches
func (bc *BatchCollector) GetBatchStatus() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	status := map[string]interface{}{
		"total_batches": len(bc.batches),
		"batches":       make([]map[string]interface{}, 0),
	}

	totalRequests := 0
	for _, b := range bc.batches {
		b.mu.Lock()
		timeRemaining := time.Until(b.ExpiresAt).Seconds()
		totalRequests += len(b.Requests)

		batchStatus := map[string]interface{}{
			"batch_id":               b.ID,
			"vehicle_type":           b.VehicleTypeID,
			"request_count":          len(b.Requests),
			"time_remaining_seconds": timeRemaining,
			"created_at":             b.CreatedAt,
		}
		status["batches"] = append(status["batches"].([]map[string]interface{}), batchStatus)
		b.mu.Unlock()
	}

	status["total_requests"] = totalRequests
	return status
}

// cleanupExpiredBatches removes batches that have expired without being processed
func (bc *BatchCollector) cleanupExpiredBatches() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bc.mu.Lock()
			now := time.Now()
			for batchID, b := range bc.batches {
				b.mu.Lock()
				if now.After(b.ExpiresAt) {
					// Batch expired, call the callback to process it
					// ✅ IMPORTANT: Callback happens BEFORE deletion so ProcessBatch can access requests
					if bc.onBatchExpire != nil {
						// Run callback in goroutine to avoid blocking cleanup
						// The callback will call ProcessBatch() which needs access to the batch
						go bc.onBatchExpire(batchID)
					}

					// ✅ FIXED: Now delete the batch after callback is triggered
					// This gives ProcessBatch() time to fetch requests before deletion
					// We use a small delay to ensure callback gets the batch
					go func(vid string, bid string) {
						// Small delay to allow ProcessBatch to read the batch
						time.Sleep(100 * time.Millisecond)
						bc.mu.Lock()
						defer bc.mu.Unlock()

						// Remove from tracking and batches
						delete(bc.vehicleTypeBatches, vid)
						delete(bc.batches, bid)
						logger.Warn("Expired batch removed after processing",
							"batchID", bid,
							"vehicleTypeID", vid,
						)
					}(b.VehicleTypeID, batchID)
				}
				b.mu.Unlock()
			}
			bc.mu.Unlock()

		case <-bc.done:
			return
		}
	}
}

// Stop stops the batch collector
func (bc *BatchCollector) Stop() {
	close(bc.done)
	if bc.ticker != nil {
		bc.ticker.Stop()
	}
}

// calculateCentroid calculates the geographic center of all requests
func (bc *BatchCollector) calculateCentroid(requests []dto.RideRequestInfo) dto.Location {
	if len(requests) == 0 {
		return dto.Location{}
	}

	var totalLat, totalLon float64
	for _, req := range requests {
		totalLat += req.PickupLat
		totalLon += req.PickupLon
	}

	centroidLat := totalLat / float64(len(requests))
	centroidLon := totalLon / float64(len(requests))

	return dto.Location{
		Latitude:  centroidLat,
		Longitude: centroidLon,
	}
}

// extractRequestIDs extracts ride IDs from requests
func extractRequestIDs(requests []dto.RideRequestInfo) []string {
	ids := make([]string, len(requests))
	for i, req := range requests {
		ids[i] = req.RideID
	}
	return ids
}
