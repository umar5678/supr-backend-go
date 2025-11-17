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

	// Validate coordinates
	if err := location.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		return response.BadRequest(err.Error())
	}

	now := time.Now()

	// Create location record
	locationRecord := &models.DriverLocationHistory{
		DriverID:  driverID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Heading:   req.Heading,
		Speed:     req.Speed,
		Accuracy:  req.Accuracy,
		Timestamp: now,
	}

	// Store in Redis immediately (for real-time access)
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

	// Save to database asynchronously (every 30 seconds in batch)
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

func (s *service) FindNearbyDrivers(ctx context.Context, req dto.FindNearbyDriversRequest) (*dto.NearbyDriversResponse, error) {
	req.SetDefaults()

	// Validate coordinates
	if err := location.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Try cache first
	cacheKey := fmt.Sprintf("nearby:drivers:%f:%f:%f:%s",
		req.Latitude, req.Longitude, req.RadiusKm, req.VehicleTypeID)

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

		// Calculate ETA (assuming average speed or use driver's current speed)
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

	// Cache for 10 seconds (short TTL for real-time accuracy)
	cache.SetJSON(ctx, cacheKey, result, 10*time.Second)

	logger.Info("nearby drivers found",
		"count", len(driverResponses),
		"radiusKm", req.RadiusKm,
	)

	return result, nil
}

func (s *service) StreamLocationToRider(ctx context.Context, rideID, driverID, riderID string) error {
	// Get driver's current location
	location, err := s.GetDriverLocation(ctx, driverID)
	if err != nil {
		return err
	}

	// Send via WebSocket to rider using the new utility
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

// New method for continuous location streaming
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
				// Continue streaming despite errors
			}
		}
	}
}

// Enhanced location update that also streams to rider if there's an active ride
func (s *service) UpdateDriverLocationWithStreaming(ctx context.Context, driverID string, req dto.UpdateLocationRequest, activeRideID, riderID string) error {
	// Update the location first
	if err := s.UpdateDriverLocation(ctx, driverID, req); err != nil {
		return err
	}

	// If there's an active ride, stream to rider
	if activeRideID != "" && riderID != "" {
		go s.StreamLocationToRider(ctx, activeRideID, driverID, riderID)
	}

	return nil
}
