package tracking

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/tracking/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/location"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
	websocketutil "github.com/umar5678/go-backend/internal/websocket/websocketutils"
)

type Service interface {
	UpdateDriverLocation(ctx context.Context, driverID string, req dto.UpdateLocationRequest) error
	GetDriverLocation(ctx context.Context, driverID string) (*dto.LocationResponse, error)
	FindNearbyDrivers(ctx context.Context, req dto.FindNearbyDriversRequest) (*dto.NearbyDriversResponse, error)
	StreamLocationToRider(ctx context.Context, rideID, driverID, riderID string) error

	// Polyline features
	// GeneratePolyline(ctx context.Context, driverID string, from, to time.Time) (string, error)
	// GetRidePolyline(ctx context.Context, rideID string) (string, error)
	// StreamPolylineToRider(ctx context.Context, rideID, driverID, riderID string, interval time.Duration)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) UpdateDriverLocation(ctx context.Context, driverID string, req dto.UpdateLocationRequest) error {
	if err := req.Validate(); err != nil {
		return response.BadRequest(err.Error())
	}

	if err := location.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		return response.BadRequest(err.Error())
	}

	now := time.Now()

	locationRecord := &models.DriverLocation{
		DriverID:  driverID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Heading:   req.Heading,
		Speed:     req.Speed,
		Accuracy:  req.Accuracy,
		Timestamp: now,
	}

	// Store in Redis immediately
	locationKey := fmt.Sprintf("driver:location:%s", driverID)
	locationData := map[string]interface{}{
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"heading":   req.Heading,
		"speed":     req.Speed,
		"accuracy":  req.Accuracy,
		"timestamp": now.Unix(),
	}

	if err := cache.SetJSON(ctx, locationKey, locationData, 30*time.Second); err != nil {
		logger.Error("failed to cache driver location", "error", err, "driverID", driverID)
	}

	// Save to database asynchronously
	go func() {
		bgCtx := context.Background()
		if err := s.repo.SaveLocation(bgCtx, locationRecord); err != nil {
			logger.Error("failed to save location to database", "error", err, "driverID", driverID)
		}
	}()

	// Refresh online status TTL
	onlineKey := fmt.Sprintf("driver:online:%s", driverID)
	cache.Set(ctx, onlineKey, "true", 5*time.Minute)

	logger.Debug("driver location updated",
		"driverID", driverID,
		"lat", req.Latitude,
		"lng", req.Longitude,
	)

	return nil
}

func (s *service) GetDriverLocation(ctx context.Context, driverID string) (*dto.LocationResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("driver:location:%s", driverID)
	var cached map[string]interface{}

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		return &dto.LocationResponse{
			Latitude:  cached["latitude"].(float64),
			Longitude: cached["longitude"].(float64),
			Heading:   int(cached["heading"].(float64)),
			Speed:     cached["speed"].(float64),
			Accuracy:  cached["accuracy"].(float64),
			Timestamp: time.Unix(int64(cached["timestamp"].(float64)), 0),
		}, nil
	}

	// Get from database
	location, err := s.repo.GetDriverLocation(ctx, driverID)
	if err != nil {
		return nil, response.NotFoundError("Driver location")
	}

	return &dto.LocationResponse{
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Heading:   location.Heading,
		Speed:     location.Speed,
		Accuracy:  location.Accuracy,
		Timestamp: location.Timestamp,
	}, nil
}

// CRITICAL FIX: Implement OnlyAvailable filter
func (s *service) FindNearbyDrivers(ctx context.Context, req dto.FindNearbyDriversRequest) (*dto.NearbyDriversResponse, error) {
	req.SetDefaults()

	if err := location.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Try cache first (but only if not filtering by availability)
	cacheKey := fmt.Sprintf("nearby:drivers:%f:%f:%f:%s:%v",
		req.Latitude, req.Longitude, req.RadiusKm, req.VehicleTypeID, req.OnlyAvailable)

	var cached dto.NearbyDriversResponse
	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		logger.Debug("nearby drivers cache hit", "lat", req.Latitude, "lng", req.Longitude)
		return &cached, nil
	}

	// Find nearby drivers from database
	drivers, err := s.repo.FindNearbyDrivers(
		ctx,
		req.Latitude,
		req.Longitude,
		req.RadiusKm,
		req.VehicleTypeID,
		req.Limit,
	)

	if err != nil {
		logger.Error("failed to find nearby drivers", "error", err)
		return nil, response.InternalServerError("Failed to find drivers", err)
	}

	// Build response
	searchPoint := location.Point{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	driverResponses := make([]dto.DriverLocationResponse, 0, len(drivers))

	for _, driver := range drivers {
		// CRITICAL: Filter by availability if requested
		if req.OnlyAvailable {
			// Check if driver is truly available (not busy with another ride)
			isAvailable, err := s.isDriverAvailable(ctx, driver.ID)
			if err != nil || !isAvailable {
				logger.Debug("skipping busy driver", "driverID", driver.ID)
				continue
			}
		}

		// Get driver's current location from Redis
		driverLoc, err := s.GetDriverLocation(ctx, driver.ID)
		if err != nil {
			continue // Skip if location not available
		}

		// Calculate distance
		driverPoint := location.Point{
			Latitude:  driverLoc.Latitude,
			Longitude: driverLoc.Longitude,
		}
		distance := location.CalculateDistance(searchPoint, driverPoint)

		// Calculate ETA
		speed := driverLoc.Speed
		if speed == 0 {
			speed = 40 // Default 40 km/h
		}
		eta := location.CalculateETA(distance, speed)

		driverResponses = append(driverResponses, dto.DriverLocationResponse{
			DriverID: driver.ID,
			Location: *driverLoc,
			Distance: distance,
			ETA:      eta,
		})
	}

	result := &dto.NearbyDriversResponse{
		SearchLocation: dto.LocationResponse{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Timestamp: time.Now(),
		},
		RadiusKm: req.RadiusKm,
		Drivers:  driverResponses,
		Count:    len(driverResponses),
	}

	// Cache for 5 seconds (short TTL for availability accuracy)
	cache.SetJSON(ctx, cacheKey, result, 5*time.Second)

	logger.Info("nearby drivers found",
		"count", len(driverResponses),
		"radiusKm", req.RadiusKm,
		"onlyAvailable", req.OnlyAvailable,
	)

	return result, nil
}

// Helper: Check if driver is truly available (not on an active ride)
func (s *service) isDriverAvailable(ctx context.Context, driverID string) (bool, error) {
	// Check Redis first for quick lookup
	busyKey := fmt.Sprintf("driver:busy:%s", driverID)
	isBusy, err := cache.Get(ctx, busyKey)
	if err == nil && isBusy == "true" {
		return false, nil
	}

	// Assume available if we can't check (optimistic)
	return true, nil
}

func (s *service) StreamLocationToRider(ctx context.Context, rideID, driverID, riderID string) error {
	location, err := s.GetDriverLocation(ctx, driverID)
	if err != nil {
		return err
	}

	locationData := map[string]interface{}{
		"rideId":   rideID,
		"driverId": driverID,
		"location": map[string]interface{}{
			"latitude":  location.Latitude,
			"longitude": location.Longitude,
			"heading":   location.Heading,
			"speed":     location.Speed,
			"accuracy":  location.Accuracy,
			"timestamp": location.Timestamp,
		},
		"timestamp": time.Now().UTC(),
	}

	if err := websocketutil.SendRideLocationUpdate(riderID, locationData); err != nil {
		logger.Error("failed to send driver location update to rider",
			"error", err,
			"riderID", riderID,
			"driverID", driverID,
			"rideID", rideID,
		)
		return err
	}

	logger.Debug("driver location streamed to rider",
		"rideID", rideID,
		"driverID", driverID,
		"riderID", riderID,
	)

	return nil
}

// Continuous location streaming
func (s *service) StartLocationStreaming(ctx context.Context, rideID, driverID, riderID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("location streaming stopped", "rideID", rideID)
			return
		case <-ticker.C:
			if err := s.StreamLocationToRider(ctx, rideID, driverID, riderID); err != nil {
				logger.Warn("failed to stream location",
					"error", err,
					"rideID", rideID,
				)
			}
		}
	}
}

// Enhanced location update with streaming
func (s *service) UpdateDriverLocationWithStreaming(ctx context.Context, driverID string, req dto.UpdateLocationRequest, activeRideID, riderID string) error {
	if err := s.UpdateDriverLocation(ctx, driverID, req); err != nil {
		return err
	}

	if activeRideID != "" && riderID != "" {
		go s.StreamLocationToRider(ctx, activeRideID, driverID, riderID)
	}

	return nil
}

// ============================================================================
// POLYLINE FEATURES
// ============================================================================

// // GeneratePolyline creates an encoded polyline from location history
// func (s *service) GeneratePolyline(ctx context.Context, driverID string, from, to time.Time) (string, error) {
// 	history, err := s.repo.GetLocationHistory(ctx, driverID, from, to, 0)
// 	if err != nil {
// 		return "", err
// 	}

// 	if len(history) == 0 {
// 		return "", response.NotFoundError("No location history found")
// 	}

// 	points := make([]location.Point, len(history))
// 	for i, loc := range history {
// 		points[i] = location.Point{
// 			Latitude:  loc.Latitude,
// 			Longitude: loc.Longitude,
// 		}
// 	}

// 	polyline := location.EncodePolyline(points)

// 	logger.Info("polyline generated",
// 		"driverID", driverID,
// 		"points", len(points),
// 		"from", from,
// 		"to", to,
// 	)
// 	points := make([]location.Point, len(allHistory))
// 	for i, loc := range allHistory {
// 		points[i] = location.Point{
// 			Latitude:  loc.Latitude,
// 			Longitude: loc.Longitude,
// 		}
// 	}

// 	polyline := location.EncodePolyline(points)
// 	cache.SetJSON(ctx, cacheKey, polyline, 30*time.Second)

// 	return polyline, nil
// }

// // GetRidePolyline retrieves polyline for an active/completed ride
// func (s *service) GetRidePolyline(ctx context.Context, rideID string) (string, error) {
// 	// Get ride details from cache first
// 	cacheKey := fmt.Sprintf("ride:polyline:%s", rideID)
// 	var cachedPolyline string
// 	if err := cache.Get(ctx, cacheKey); err == nil {
// 		return cachedPolyline, nil
// 	}

// 	// Fetch ride to get driver and timestamps
// 	ride, err := s.repo.GetRideForPolyline(ctx, rideID)
// 	if err != nil {
// 		return "", response.NotFoundError("Ride")
// 	}

// 	if ride.DriverID == nil {
// 		return "", response.BadRequest("No driver assigned to this ride")
// 	}

// 	// Determine time range
// 	from := ride.AcceptedAt
// 	if from == nil {
// 		from = &ride.RequestedAt
// 	}

// 	to := ride.CompletedAt
// 	if to == nil {
// 		now := time.Now()
// 		to = &now
// 	}

// 	polyline, err := s.GeneratePolyline(ctx, *ride.DriverID, *from, *to)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Cache completed ride polylines for longer
// 	if ride.Status == "completed" {
// 		cache.Set(ctx, cacheKey, polyline, 24*time.Hour)
// 	} else {
// 		cache.Set(ctx, cacheKey, polyline, 30*time.Second)
// 	}

// 	return polyline, nil
// }

// // StreamPolylineToRider pushes updated polyline during trip
// func (s *service) StreamPolylineToRider(ctx context.Context, rideID, driverID, riderID string, interval time.Duration) {
// 	ticker := time.NewTicker(interval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			logger.Info("polyline streaming stopped", "rideID", rideID)
// 			return

// 		case <-ticker.C:
// 			polyline, err := s.GetRidePolyline(ctx, rideID)
// 			if err != nil {
// 				logger.Warn("failed to generate polyline", "error", err, "rideID", rideID)
// 				continue
// 			}

// 			data := map[string]interface{}{
// 				"rideId":    rideID,
// 				"polyline":  polyline,
// 				"timestamp": time.Now().UTC(),
// 			}

// 			if err := websocketutil.SendRideLocationUpdate(riderID, data); err != nil {
// 				logger.Warn("failed to send polyline update", "error", err, "rideID", rideID)
// 			}
// 		}
// 	}
// }
