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

	go func() {
		bgCtx := context.Background()
		if err := s.repo.SaveLocation(bgCtx, locationRecord); err != nil {
			logger.Error("failed to save location to database", "error", err, "driverID", driverID)
		}
	}()

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

	cacheKey := fmt.Sprintf("driver:location:%s", driverID)
	var cached map[string]interface{}

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil && cached != nil {

		lat, latOk := cached["latitude"].(float64)
		lon, lonOk := cached["longitude"].(float64)

		if !latOk || !lonOk {

			logger.Warn("invalid cache data for driver location", "driverID", driverID)
		} else {

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
	logger.Info("GetDriverProfileID CALLED", "userID", userID)

	cacheKey := fmt.Sprintf("user:driver_profile:%s", userID)
	logger.Info("Checking cache", "cacheKey", cacheKey)

	profileID, err := cache.Get(ctx, cacheKey)
	if err == nil && profileID != "" {
		logger.Info("Profile ID found in CACHE", "profileID", profileID)
		return profileID, nil
	}

	logger.Info("Not in cache, querying DATABASE")

	var profile struct {
		ID string
	}

	err = s.repo.GetDB().
		Table("driver_profiles").
		Select("id").
		Where("user_id = ?", userID).
		First(&profile).Error

	if err != nil {
		logger.Error("--------- ---------- - - Driver profile NOT found in database",
			"userID", userID,
			"error", err,
		)
		return "", err
	}

	logger.Info("---------- -------- Driver profile found in DATABASE",
		"userID", userID,
		"profileID", profile.ID,
	)

	cache.Set(ctx, cacheKey, profile.ID, 1*time.Hour)
	return profile.ID, nil
}

func (s *service) GetDriverActiveRide(ctx context.Context, driverID string) (string, string, error) {
	logger.Info("GetDriverActiveRide CALLED", "driverProfileID", driverID)

	cacheKey := fmt.Sprintf("driver:active:ride:%s", driverID)
	logger.Info("Checking cache", "cacheKey", cacheKey)

	var rideData map[string]string

	err := cache.GetJSON(ctx, cacheKey, &rideData)
	if err == nil && rideData != nil {
		logger.Info("---- --- ---  Active ride found in CACHE",
			"rideID", rideData["rideID"],
			"riderID", rideData["riderID"],
		)
		return rideData["rideID"], rideData["riderID"], nil
	}

	logger.Warn("Not in cache, querying DATABASE", "cacheError", err)

	var ride struct {
		ID      string
		RiderID string
	}

	query := s.repo.GetDB().
		Table("rides").
		Select("id, rider_id").
		Where("driver_id = ?", driverID).
		Where("status IN (?)", []string{"accepted", "started"})

	logger.Info("Executing query", "driverID", driverID)

	err = query.First(&ride).Error

	if err != nil {
		logger.Error("============ No active ride in DATABASE",
			"driverProfileID", driverID,
			"error", err,
		)
		return "", "", err
	}

	logger.Info("Active ride found in DATABASE",
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
	logger.Info("UpdateDriverLocationWithStreaming CALLED",
		"driverProfileID", driverID,
		"rideID", activeRideID,
		"riderUserID", riderID,
	)

	logger.Info("Updating driver location in database...")
	if err := s.UpdateDriverLocation(ctx, driverID, req); err != nil {
		logger.Error("========================= UpdateDriverLocation FAILED", "error", err)
		return err
	}
	logger.Info("Driver location updated in database")

	if activeRideID == "" {
		logger.Error("========================= Empty activeRideID, cannot stream")
		return nil
	}
	if riderID == "" {
		logger.Error("========================  Empty riderID, cannot stream")
		return nil
	}

	logger.Info("Starting goroutine for StreamLocationToRider")
	go func() {
		logger.Info("Goroutine STARTED for location streaming")
		if err := s.StreamLocationToRider(context.Background(), activeRideID, driverID, riderID); err != nil {
			logger.Error("===================  StreamLocationToRider error in goroutine ==== ", "error", err)
		}
	}()

	logger.Info("UpdateDriverLocationWithStreaming completed (goroutine spawned)")
	return nil
}

func (s *service) StreamLocationToRider(ctx context.Context, rideID, driverID, riderID string) error {
	logger.Info("===========================")
	logger.Info("StreamLocationToRider STARTED",
		"rideID", rideID,
		"driverProfileID", driverID,
		"riderUserID", riderID,
	)

	logger.Info("Getting driver location...")
	location, err := s.GetDriverLocation(ctx, driverID)
	if err != nil {
		logger.Error("GetDriverLocation FAILED",
			"error", err,
			"driverID", driverID,
		)
		return err
	}

	logger.Info("Driver location retrieved",
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

	logger.Info("Calling SendRideLocationUpdate",
		"riderUserID", riderID,
		"locationData", locationData,
	)

	if err := websocketutil.SendRideLocationUpdate(riderID, locationData); err != nil {
		logger.Error("============== SendRideLocationUpdate FAILED ==================",
			"error", err,
			"riderUserID", riderID,
		)
		return err
	}

	logger.Info("StreamLocationToRider COMPLETED SUCCESSFULLY")
	logger.Info("============================================================")
	return nil
}

func (s *service) FindNearbyDrivers(ctx context.Context, req dto.FindNearbyDriversRequest) (*dto.NearbyDriversResponse, error) {
	req.SetDefaults()

	if err := location.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, response.BadRequest(err.Error())
	}

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

	searchPoint := location.Point{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	driverResponses := make([]dto.DriverLocationResponse, 0, len(drivers))

	for _, driver := range drivers {
		if req.OnlyAvailable {
			isAvailable, err := s.isDriverAvailable(ctx, driver.ID)
			if err != nil || !isAvailable {
				logger.Debug("skipping busy driver", "driverID", driver.ID)
				continue
			}
		}

		driverLoc, err := s.GetDriverLocation(ctx, driver.ID)
		if err != nil {
			logger.Debug("skipping driver with no location", "driverID", driver.ID)
			continue
		}

		driverPoint := location.Point{
			Latitude:  driverLoc.Latitude,
			Longitude: driverLoc.Longitude,
		}
		distance := location.CalculateDistance(searchPoint, driverPoint)

		speed := driverLoc.Speed
		if speed == 0 {
			speed = 40
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

func (s *service) isDriverAvailable(ctx context.Context, driverID string) (bool, error) {

	busyKey := fmt.Sprintf("driver:busy:%s", driverID)
	isBusy, err := cache.Get(ctx, busyKey)
	if err == nil && isBusy == "true" {
		return false, nil
	}

	return true, nil
}


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
