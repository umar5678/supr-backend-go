package services

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/drivers"
	"github.com/umar5678/go-backend/internal/modules/rides"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/location"
	"github.com/umar5678/go-backend/internal/utils/logger"
	websocketutil "github.com/umar5678/go-backend/internal/websocket/websocketutils"
)

type LocationService interface {
	UpdateDriverLocation(ctx context.Context, driverID string, lat, lng float64, heading int, speed, accuracy float64) error
	BroadcastLocationToRider(ctx context.Context, driverID, rideID string, location *models.DriverLocation) error
	GetDriverLocation(ctx context.Context, driverID string) (*models.DriverLocation, error)
	CalculateRoute(ctx context.Context, startLat, startLng, endLat, endLng float64) (*models.Route, error)
	CalculateETA(ctx context.Context, driverLat, driverLng, destLat, destLng float64) (int, error)
}

type locationService struct {
	driverRepo drivers.Repository
	rideRepo   rides.Repository
}

func NewLocationService(driverRepo drivers.Repository, rideRepo rides.Repository) LocationService {
	return &locationService{
		driverRepo: driverRepo,
		rideRepo:   rideRepo,
	}
}

func (s *locationService) UpdateDriverLocation(ctx context.Context, driverID string, lat, lng float64, heading int, speed, accuracy float64) error {
	location := &models.DriverLocation{
		DriverID:  driverID,
		Latitude:  lat,
		Longitude: lng,
		Heading:   heading,
		Speed:     speed,
		Accuracy:  accuracy,
		Timestamp: time.Now(),
	}

	if err := s.driverRepo.UpdateDriverLocation(ctx, driverID, lat, lng, heading); err != nil {
		return err
	}

	locationKey := fmt.Sprintf("driver:location:%s", driverID)
	locationData := map[string]interface{}{
		"latitude":  lat,
		"longitude": lng,
		"heading":   heading,
		"speed":     speed,
		"accuracy":  accuracy,
		"timestamp": time.Now().Unix(),
	}
	cache.SetJSON(ctx, locationKey, locationData, 30*time.Second)

	activeRide, err := s.rideRepo.FindActiveRideByDriverID(ctx, driverID)
	if err == nil && activeRide != nil {
		go s.BroadcastLocationToRider(context.Background(), driverID, activeRide.ID, location)
	}

	logger.Debug("driver location updated",
		"driverID", driverID,
		"lat", lat,
		"lng", lng,
		"speed", speed,
	)

	return nil
}

func (s *locationService) BroadcastLocationToRider(ctx context.Context, driverID, rideID string, location *models.DriverLocation) error {
	ride, err := s.rideRepo.FindRideByID(ctx, rideID)
	if err != nil {
		return err
	}

	eta, err := s.CalculateETA(ctx, location.Latitude, location.Longitude, ride.DropoffLat, ride.DropoffLon)
	if err != nil {
		eta = 0
	}

	broadcast := models.LocationBroadcast{
		DriverID:  driverID,
		RideID:    rideID,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Heading:   location.Heading,
		Speed:     location.Speed,
		Timestamp: location.Timestamp,
		ETA:       eta,
	}

	message := map[string]interface{}{
		"type":      "driver_location_update",
		"data":      broadcast,
		"timestamp": time.Now().UTC(),
	}

	if err := websocketutil.SendToUser(ride.RiderID, "driver_location_update", message); err != nil {
		logger.Error("failed to broadcast location to rider",
			"error", err,
			"riderID", ride.RiderID,
			"rideID", rideID,
		)
		return err
	}

	return nil
}
func (s *locationService) GetDriverLocation(ctx context.Context, driverID string) (*models.DriverLocation, error) {
	cacheKey := fmt.Sprintf("driver:location:%s", driverID)
	var cachedLocation models.DriverLocation

	err := cache.GetJSON(ctx, cacheKey, &cachedLocation)
	if err == nil {
		return &cachedLocation, nil
	}

	location, err := s.driverRepo.GetLatestDriverLocation(ctx, driverID)
	if err != nil {
		logger.Error("failed to get driver location from database",
			"error", err,
			"driverID", driverID,
		)
		return nil, fmt.Errorf("driver location not found")
	}

	locationData := map[string]interface{}{
		"latitude":  location.Latitude,
		"longitude": location.Longitude,
		"heading":   location.Heading,
		"speed":     location.Speed,
		"accuracy":  location.Accuracy,
		"timestamp": location.Timestamp.Unix(),
	}
	cache.SetJSON(ctx, cacheKey, locationData, 30*time.Second)

	return location, nil
}

func (s *locationService) CalculateRoute(ctx context.Context, startLat, startLng, endLat, endLng float64) (*models.Route, error) {
	return &models.Route{
		Polyline: "polyline_encoded_string",
		Distance: 5000,
		Duration: 900,
		Summary:  "Fastest route",
	}, nil
}

func (s *locationService) CalculateETA(ctx context.Context, driverLat, driverLng, destLat, destLng float64) (int, error) {
	distanceKm := location.HaversineDistance(driverLat, driverLng, destLat, destLng)
	// Convert km/h to m/s: 13.8889 m/s ≈ 50 km/h average speed
	eta := int((distanceKm / 13.8889) * 3600)
	return eta, nil
}

func calculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// DEPRECATED: Use location.HaversineDistance instead
	// This function returns km, not meters
	return location.HaversineDistance(lat1, lon1, lat2, lon2)
}
