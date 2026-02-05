package batching

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/modules/batching/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type BatchCollector struct {
	batchWindow        time.Duration     
	maxBatchSize       int               
	batches            map[string]*batch 
	vehicleTypeBatches map[string]string 
	mu                 sync.RWMutex
	ticker             *time.Ticker
	done               chan struct{}
	onBatchExpire      func(batchID string) 
}

type batch struct {
	ID            string
	Requests      []dto.RideRequestInfo
	CreatedAt     time.Time
	ExpiresAt     time.Time
	VehicleTypeID string
	mu            sync.Mutex
}

func NewBatchCollector(batchWindow time.Duration, maxBatchSize int) *BatchCollector {
	bc := &BatchCollector{
		batchWindow:        batchWindow,
		maxBatchSize:       maxBatchSize,
		batches:            make(map[string]*batch),
		vehicleTypeBatches: make(map[string]string),
		done:               make(chan struct{}),
		onBatchExpire:      nil, 
	}

	
	go bc.cleanupExpiredBatches()

	return bc
}


func (bc *BatchCollector) SetBatchExpireCallback(callback func(batchID string)) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.onBatchExpire = callback
}

func (bc *BatchCollector) AddRequest(ctx context.Context, req dto.RideRequestInfo) (string, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

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

	b.mu.Lock()
	b.Requests = append(b.Requests, req)
	requestCount := len(b.Requests)
	b.mu.Unlock()

	logger.Debug("Added request to batch",
		"batchID", batchID,
		"rideID", req.RideID,
		"requestCount", requestCount,
	)

	return batchID, nil
}

func (bc *BatchCollector) GetBatch(batchID string) (*dto.RequestBatch, error) {
	bc.mu.RLock()
	b, exists := bc.batches[batchID]
	bc.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

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

func (bc *BatchCollector) GetBatchRequests(batchID string) ([]dto.RideRequestInfo, error) {
	bc.mu.RLock()
	b, exists := bc.batches[batchID]
	bc.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	requests := make([]dto.RideRequestInfo, len(b.Requests))
	copy(requests, b.Requests)

	return requests, nil
}

func (bc *BatchCollector) CheckBatchReady(batchID string) bool {
	bc.mu.RLock()
	b, exists := bc.batches[batchID]
	bc.mu.RUnlock()

	if !exists {
		return false
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	isExpired := time.Now().After(b.ExpiresAt)
	isMaxed := len(b.Requests) >= bc.maxBatchSize

	return isExpired || isMaxed
}

func (bc *BatchCollector) CompleteBatch(batchID string) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if b, exists := bc.batches[batchID]; exists {
		delete(bc.vehicleTypeBatches, b.VehicleTypeID)
		delete(bc.batches, batchID)

		logger.Info("Completed batch",
			"batchID", batchID,
			"vehicleTypeID", b.VehicleTypeID,
		)
	}

	return nil
}

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
					if bc.onBatchExpire != nil {
						go bc.onBatchExpire(batchID)
					}

					go func(vid string, bid string) {
						time.Sleep(100 * time.Millisecond)
						bc.mu.Lock()
						defer bc.mu.Unlock()

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

func (bc *BatchCollector) Stop() {
	close(bc.done)
	if bc.ticker != nil {
		bc.ticker.Stop()
	}
}

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

func extractRequestIDs(requests []dto.RideRequestInfo) []string {
	ids := make([]string, len(requests))
	for i, req := range requests {
		ids[i] = req.RideID
	}
	return ids
}
