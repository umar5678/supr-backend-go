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
	GetDriverProfileID(ctx context.Context, userID string) (string, error)
	GetDriverActiveRide(ctx context.Context, driverID string) (rideID, riderID string, err error)
	UpdateDriverLocationWithStreaming(ctx context.Context, driverID string, req dto.UpdateLocationRequest, activeRideID, riderID string) error
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

// internal/modules/tracking/service.go

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
	if err == nil && cached != nil {
		// Safely extract values with type checking
		lat, latOk := cached["latitude"].(float64)
		lon, lonOk := cached["longitude"].(float64)

		if !latOk || !lonOk {
			// Invalid cache data, fall through to database
			logger.Warn("invalid cache data for driver location", "driverID", driverID)
		} else {
			// Safe extraction with defaults for optional fields
			heading := 0
			if h, ok := cached["heading"].(float64); ok {
				heading = int(h)
			}

			speed := 0.0
			if s, ok := cached["speed"].(float64); ok {
				speed = s
			}

			accuracy := 0.0
			if a, ok := cached["accuracy"].(float64); ok {
				accuracy = a
			}

			timestamp := time.Now()
			if ts, ok := cached["timestamp"].(float64); ok {
				timestamp = time.Unix(int64(ts), 0)
			}

			return &dto.LocationResponse{
				Latitude:  lat,
				Longitude: lon,
				Heading:   heading,
				Speed:     speed,
				Accuracy:  accuracy,
				Timestamp: timestamp,
			}, nil
		}
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

func (s *service) GetDriverProfileID(ctx context.Context, userID string) (string, error) {
	logger.Info("🔍 GetDriverProfileID CALLED", "userID", userID)

	cacheKey := fmt.Sprintf("user:driver_profile:%s", userID)
	logger.Info("📦 Checking cache", "cacheKey", cacheKey)

	profileID, err := cache.Get(ctx, cacheKey)
	if err == nil && profileID != "" {
		logger.Info("✅ Profile ID found in CACHE", "profileID", profileID)
		return profileID, nil
	}

	logger.Info("⚠️ Not in cache, querying DATABASE")

	var profile struct {
		ID string
	}

	err = s.repo.GetDB().
		Table("driver_profiles").
		Select("id").
		Where("user_id = ?", userID).
		First(&profile).Error

	if err != nil {
		logger.Error("--------- ---------- - - ❌ Driver profile NOT found in database",
			"userID", userID,
			"error", err,
		)
		return "", err
	}

	logger.Info("---------- -------- ✅ Driver profile found in DATABASE",
		"userID", userID,
		"profileID", profile.ID,
	)

	cache.Set(ctx, cacheKey, profile.ID, 1*time.Hour)
	return profile.ID, nil
}

func (s *service) GetDriverActiveRide(ctx context.Context, driverID string) (string, string, error) {
	logger.Info("🔍 GetDriverActiveRide CALLED", "driverProfileID", driverID)

	cacheKey := fmt.Sprintf("driver:active:ride:%s", driverID)
	logger.Info("📦 Checking cache", "cacheKey", cacheKey)

	var rideData map[string]string

	err := cache.GetJSON(ctx, cacheKey, &rideData)
	if err == nil && rideData != nil {
		logger.Info("---- --- ---  ✅ Active ride found in CACHE",
			"rideID", rideData["rideID"],
			"riderID", rideData["riderID"],
		)
		return rideData["rideID"], rideData["riderID"], nil
	}

	logger.Warn("⚠️ Not in cache, querying DATABASE", "cacheError", err)

	var ride struct {
		ID      string
		RiderID string
	}

	query := s.repo.GetDB().
		Table("rides").
		Select("id, rider_id").
		Where("driver_id = ?", driverID).
		Where("status IN (?)", []string{"accepted", "started"})

	logger.Info("📊 Executing query", "driverID", driverID)

	err = query.First(&ride).Error

	if err != nil {
		logger.Error("============ ❌ No active ride in DATABASE",
			"driverProfileID", driverID,
			"error", err,
		)
		return "", "", err
	}

	logger.Info("✅ Active ride found in DATABASE",
		"rideID", ride.ID,
		"riderID", ride.RiderID,
	)

	cache.SetJSON(ctx, cacheKey, map[string]string{
		"rideID":  ride.ID,
		"riderID": ride.RiderID,
	}, 30*time.Minute)

	return ride.ID, ride.RiderID, nil
}

func (s *service) UpdateDriverLocationWithStreaming(ctx context.Context, driverID string, req dto.UpdateLocationRequest, activeRideID, riderID string) error {
	logger.Info("🚗 UpdateDriverLocationWithStreaming CALLED",
		"driverProfileID", driverID,
		"rideID", activeRideID,
		"riderUserID", riderID,
	)

	// Update location first
	logger.Info("📍 Updating driver location in database...")
	if err := s.UpdateDriverLocation(ctx, driverID, req); err != nil {
		logger.Error("========================= ❌ UpdateDriverLocation FAILED", "error", err)
		return err
	}
	logger.Info("✅ Driver location updated in database")

	// Validate IDs
	if activeRideID == "" {
		logger.Error("========================= ❌ Empty activeRideID, cannot stream")
		return nil
	}
	if riderID == "" {
		logger.Error("========================  ❌ Empty riderID, cannot stream")
		return nil
	}

	// Stream to rider (non-blocking)
	logger.Info("📡 Starting goroutine for StreamLocationToRider")
	go func() {
		logger.Info("🎯 Goroutine STARTED for location streaming")
		if err := s.StreamLocationToRider(context.Background(), activeRideID, driverID, riderID); err != nil {
			logger.Error("===================  ❌ StreamLocationToRider error in goroutine ==== ", "error", err)
		}
	}()

	logger.Info("✅ UpdateDriverLocationWithStreaming completed (goroutine spawned)")
	return nil
}

func (s *service) StreamLocationToRider(ctx context.Context, rideID, driverID, riderID string) error {
	logger.Info("===========================")
	logger.Info("🎯 StreamLocationToRider STARTED",
		"rideID", rideID,
		"driverProfileID", driverID,
		"riderUserID", riderID,
	)

	logger.Info("📍 Getting driver location...")
	location, err := s.GetDriverLocation(ctx, driverID)
	if err != nil {
		logger.Error("❌ GetDriverLocation FAILED",
			"error", err,
			"driverID", driverID,
		)
		return err
	}

	logger.Info("✅ Driver location retrieved",
		"lat", location.Latitude,
		"lng", location.Longitude,
	)

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

	logger.Info("📤 Calling SendRideLocationUpdate",
		"riderUserID", riderID,
		"locationData", locationData,
	)

	if err := websocketutil.SendRideLocationUpdate(riderID, locationData); err != nil {
		logger.Error("============== ❌ SendRideLocationUpdate FAILED ==================",
			"error", err,
			"riderUserID", riderID,
		)
		return err
	}

	logger.Info("✅ StreamLocationToRider COMPLETED SUCCESSFULLY")
	logger.Info("============================================================")
	return nil
}

func (s *service) FindNearbyDrivers(ctx context.Context, req dto.FindNearbyDriversRequest) (*dto.NearbyDriversResponse, error) {
	req.SetDefaults()

	if err := location.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Only use cache if NOT filtering by availability (since that changes frequently)
	var cached dto.NearbyDriversResponse
	if !req.OnlyAvailable {
		cacheKey := fmt.Sprintf("nearby:drivers:%f:%f:%f:%s",
			req.Latitude, req.Longitude, req.RadiusKm, req.VehicleTypeID)

		err := cache.GetJSON(ctx, cacheKey, &cached)
		if err == nil {
			logger.Debug("nearby drivers cache hit", "lat", req.Latitude, "lng", req.Longitude)
			return &cached, nil
		}
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
		// Filter by availability if requested
		if req.OnlyAvailable {
			isAvailable, err := s.isDriverAvailable(ctx, driver.ID)
			if err != nil || !isAvailable {
				logger.Debug("skipping busy driver", "driverID", driver.ID)
				continue
			}
		}

		// Get driver's current location from Redis (with error handling)
		driverLoc, err := s.GetDriverLocation(ctx, driver.ID)
		if err != nil {
			logger.Debug("skipping driver with no location", "driverID", driver.ID)
			continue
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

	// Only cache if NOT filtering by availability
	if !req.OnlyAvailable {
		cacheKey := fmt.Sprintf("nearby:drivers:%f:%f:%f:%s",
			req.Latitude, req.Longitude, req.RadiusKm, req.VehicleTypeID)
		cache.SetJSON(ctx, cacheKey, result, 10*time.Second)
	}

	logger.Info("nearby drivers found",
		"count", len(driverResponses),
		"radiusKm", req.RadiusKm,
		"onlyAvailable", req.OnlyAvailable,
	)

	return result, nil
}

// ✅ Enhanced GetDriverActiveRide
// func (s *service) GetDriverActiveRide(ctx context.Context, driverID string) (string, string, error) {
// 	cacheKey := fmt.Sprintf("driver:active:ride:%s", driverID)

// 	logger.Info("🔍 Looking for active ride",
// 		"driverProfileID", driverID,
// 		"cacheKey", cacheKey,
// 	)

// 	var rideData map[string]string

// 	err := cache.GetJSON(ctx, cacheKey, &rideData)
// 	if err == nil && rideData != nil {
// 		logger.Info("✅ Active ride found in CACHE",
// 			"driverProfileID", driverID,
// 			"rideID", rideData["rideID"],
// 			"riderUserID", rideData["riderID"],
// 		)
// 		return rideData["rideID"], rideData["riderID"], nil
// 	}

// 	logger.Warn("⚠️ No active ride in cache, checking DATABASE",
// 		"driverProfileID", driverID,
// 		"cacheError", err,
// 	)

// 	// Query database
// 	var ride struct {
// 		ID      string
// 		RiderID string
// 	}

// 	err = s.repo.GetDB().
// 		Table("rides").
// 		Select("id, rider_id").
// 		Where("driver_id = ?", driverID).
// 		Where("status IN (?)", []string{"accepted", "started"}).
// 		First(&ride).Error

// 	if err != nil {
// 		logger.Warn("❌ No active ride in DATABASE",
// 			"driverProfileID", driverID,
// 			"error", err,
// 		)
// 		return "", "", err
// 	}

// 	logger.Info("✅ Active ride found in DATABASE",
// 		"driverProfileID", driverID,
// 		"rideID", ride.ID,
// 		"riderUserID", ride.RiderID,
// 	)

// 	// Cache for 30 minutes
// 	cache.SetJSON(ctx, cacheKey, map[string]string{
// 		"rideID":  ride.ID,
// 		"riderID": ride.RiderID,
// 	}, 30*time.Minute)

// 	return ride.ID, ride.RiderID, nil
// }

// // ✅ Enhanced UpdateDriverLocationWithStreaming
// func (s *service) UpdateDriverLocationWithStreaming(ctx context.Context, driverID string, req dto.UpdateLocationRequest, activeRideID, riderID string) error {
// 	logger.Info("🚗 Updating location with streaming",
// 		"driverProfileID", driverID,
// 		"rideID", activeRideID,
// 		"riderUserID", riderID,
// 	)

// 	// Update location first
// 	if err := s.UpdateDriverLocation(ctx, driverID, req); err != nil {
// 		logger.Error("❌ Failed to update driver location", "error", err)
// 		return err
// 	}

// 	// Stream to rider (non-blocking)
// 	if activeRideID != "" && riderID != "" {
// 		logger.Info("📡 Starting location stream to rider",
// 			"riderUserID", riderID,
// 			"rideID", activeRideID,
// 		)

// 		go s.StreamLocationToRider(ctx, activeRideID, driverID, riderID)
// 	} else {
// 		logger.Warn("⚠️ Empty rideID or riderID, skipping stream",
// 			"rideID", activeRideID,
// 			"riderID", riderID,
// 		)
// 	}

// 	return nil
// }

// // ✅ Enhanced StreamLocationToRider
// func (s *service) StreamLocationToRider(ctx context.Context, rideID, driverID, riderID string) error {
// 	logger.Info("🎯 StreamLocationToRider called",
// 		"rideID", rideID,
// 		"driverProfileID", driverID,
// 		"riderUserID", riderID,
// 	)

// 	location, err := s.GetDriverLocation(ctx, driverID)
// 	if err != nil {
// 		logger.Error("❌ Failed to get driver location",
// 			"error", err,
// 			"driverID", driverID,
// 		)
// 		return err
// 	}

// 	locationData := map[string]interface{}{
// 		"rideId":   rideID,
// 		"driverId": driverID,
// 		"location": map[string]interface{}{
// 			"latitude":  location.Latitude,
// 			"longitude": location.Longitude,
// 			"heading":   location.Heading,
// 			"speed":     location.Speed,
// 			"accuracy":  location.Accuracy,
// 			"timestamp": location.Timestamp,
// 		},
// 		"timestamp": time.Now().UTC(),
// 	}

// 	logger.Info("📤 Sending location to rider via WebSocket",
// 		"riderUserID", riderID,
// 		"lat", location.Latitude,
// 		"lng", location.Longitude,
// 		"rideID", rideID,
// 	)

// 	if err := websocketutil.SendRideLocationUpdate(riderID, locationData); err != nil {
// 		logger.Error("❌ WebSocket send FAILED",
// 			"error", err,
// 			"riderUserID", riderID,
// 			"rideID", rideID,
// 		)
// 		return err
// 	}

// 	logger.Info("✅ Location successfully sent to rider",
// 		"riderUserID", riderID,
// 		"rideID", rideID,
// 	)

// 	return nil
// }

// func (s *service) GetDriverActiveRide(ctx context.Context, driverID string) (string, string, error) {
// 	// Check cache first
// 	cacheKey := fmt.Sprintf("driver:active:ride:%s", driverID)

// 	logger.Debug("🔍 Checking for active ride",
// 		"cacheKey", cacheKey,
// 		"driverID", driverID,
// 	)

// 	var rideData map[string]string

// 	err := cache.GetJSON(ctx, cacheKey, &rideData)
// 	if err == nil && rideData != nil {
// 		logger.Info("✅ Active ride found in cache",
// 			"driverID", driverID,
// 			"rideID", rideData["rideID"],
// 			"riderID", rideData["riderID"],
// 		)
// 		return rideData["rideID"], rideData["riderID"], nil
// 	}

// 	logger.Debug("❌ No active ride in cache, checking database",
// 		"driverID", driverID,
// 		"cacheError", err,
// 	)

// 	// Query database
// 	var ride struct {
// 		ID      string
// 		RiderID string
// 	}

// 	err = s.repo.GetDB().
// 		Table("rides").
// 		Select("id, rider_id").
// 		Where("driver_id = ?", driverID).
// 		Where("status IN (?)", []string{"accepted", "started"}).
// 		First(&ride).Error

// 	if err != nil {
// 		logger.Debug("❌ No active ride in database",
// 			"driverID", driverID,
// 			"error", err,
// 		)
// 		return "", "", err
// 	}

// 	logger.Info("✅ Active ride found in database",
// 		"driverID", driverID,
// 		"rideID", ride.ID,
// 		"riderID", ride.RiderID,
// 	)

// 	// Cache for 30 minutes
// 	cache.SetJSON(ctx, cacheKey, map[string]string{
// 		"rideID":  ride.ID,
// 		"riderID": ride.RiderID,
// 	}, 30*time.Minute)

// 	return ride.ID, ride.RiderID, nil
// }
// func (s *service) UpdateDriverLocationWithStreaming(ctx context.Context, driverID string, req dto.UpdateLocationRequest, activeRideID, riderID string) error {
// 	// Update location first
// 	if err := s.UpdateDriverLocation(ctx, driverID, req); err != nil {
// 		return err
// 	}

// 	// ✅ Validate and verify rider user ID exists
// 	if activeRideID != "" && riderID != "" {
// 		// Verify rider user exists (optional but recommended)
// 		var userExists bool
// 		err := s.repo.GetDB().
// 			Table("users").
// 			Select("EXISTS(SELECT 1 FROM users WHERE id = ?)", riderID).
// 			Scan(&userExists).Error

// 		if err != nil || !userExists {
// 			logger.Warn("rider user not found, skipping location stream",
// 				"riderID", riderID,
// 				"rideID", activeRideID,
// 				"error", err,
// 			)
// 			return nil // Don't fail the location update
// 		}

// 		logger.Debug("streaming location to verified rider user",
// 			"riderUserID", riderID,
// 			"rideID", activeRideID,
// 			"driverID", driverID,
// 		)

// 		// Stream to rider (non-blocking)
// 		go s.StreamLocationToRider(ctx, activeRideID, driverID, riderID)
// 	}

// 	return nil
// }

// func (s *service) UpdateDriverLocationWithStreaming(ctx context.Context, driverID string, req dto.UpdateLocationRequest, activeRideID, riderID string) error {
// 	// Update location first
// 	if err := s.UpdateDriverLocation(ctx, driverID, req); err != nil {
// 		return err
// 	}

// 	// fetch user form rider id then send data to that user Id

// 	// Stream to rider (non-blocking)
// 	if activeRideID != "" && riderID != "" {
// 		go s.StreamLocationToRider(ctx, activeRideID, driverID, riderID)
// 	}

// 	return nil
// }

// ✅ NEW: Get driver profile ID from user ID
// func (s *service) GetDriverProfileID(ctx context.Context, userID string) (string, error) {
// 	// Check cache first
// 	cacheKey := fmt.Sprintf("user:driver_profile:%s", userID)
// 	var profileID string

// 	profileID, err := cache.Get(ctx, cacheKey)
// 	if err == nil && profileID != "" {
// 		return profileID, nil
// 	}
// 	// Query database
// 	var profile struct {
// 		ID string
// 	}

// 	err = s.repo.GetDB().
// 		Table("driver_profiles").
// 		Select("id").
// 		Where("user_id = ?", userID).
// 		First(&profile).Error

// 	if err != nil {
// 		return "", err
// 	}

// 	// Cache for 1 hour (profile ID doesn't change)
// 	cache.Set(ctx, cacheKey, profile.ID, 1*time.Hour)

// 	return profile.ID, nil
// }

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

// func (s *service) StreamLocationToRider(ctx context.Context, rideID, driverID, riderID string) error {
// 	location, err := s.GetDriverLocation(ctx, driverID)
// 	if err != nil {
// 		logger.Error("failed to get driver location for streaming",
// 			"error", err,
// 			"driverID", driverID,
// 			"rideID", rideID,
// 		)
// 		return err
// 	}

// 	locationData := map[string]interface{}{
// 		"rideId":   rideID,
// 		"driverId": driverID,
// 		"location": map[string]interface{}{
// 			"latitude":  location.Latitude,
// 			"longitude": location.Longitude,
// 			"heading":   location.Heading,
// 			"speed":     location.Speed,
// 			"accuracy":  location.Accuracy,
// 			"timestamp": location.Timestamp,
// 		},
// 		"timestamp": time.Now().UTC(),
// 	}

// 	logger.Info("🚗 Streaming location to rider",
// 		"rideID", rideID,
// 		"driverID", driverID,
// 		"riderID", riderID,
// 		"lat", location.Latitude,
// 		"lng", location.Longitude,
// 	)

// 	if err := websocketutil.SendRideLocationUpdate(riderID, locationData); err != nil {
// 		logger.Error("❌ failed to send driver location update to rider",
// 			"error", err,
// 			"riderID", riderID,
// 			"driverID", driverID,
// 			"rideID", rideID,
// 		)
// 		return err
// 	}

// 	logger.Info("✅ Driver location successfully streamed to rider",
// 		"rideID", rideID,
// 		"driverID", driverID,
// 		"riderID", riderID,
// 	)

// 	return nil
// }

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
