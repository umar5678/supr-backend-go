// internal/modules/pricing/service.go
package pricing

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/pricing/dto"
	vehiclesrepo "github.com/umar5678/go-backend/internal/modules/vehicles"
	"github.com/umar5678/go-backend/internal/services/cache"

	"github.com/umar5678/go-backend/internal/utils/location"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	GetFareEstimate(ctx context.Context, req dto.FareEstimateRequest) (*dto.FareEstimateResponse, error)
	CalculateActualFare(ctx context.Context, req dto.CalculateActualFareRequest) (*dto.FareEstimateResponse, error)
	GetSurgeMultiplier(ctx context.Context, lat, lon float64) (float64, error)
	GetActiveSurgeZones(ctx context.Context) ([]*dto.SurgeZoneResponse, error)
	GetFareBreakdown(ctx context.Context, estimate *models.FareEstimate) *dto.FareBreakdownResponse
}

type service struct {
	repo         Repository
	vehiclesRepo vehiclesrepo.Repository
	calculator   *FareCalculator
	surgeManager *SurgeManager
}

func NewService(repo Repository, vehiclesRepo vehiclesrepo.Repository) Service {
	return &service{
		repo:         repo,
		vehiclesRepo: vehiclesRepo,
		calculator:   NewFareCalculator(),
		surgeManager: NewSurgeManager(repo),
	}
}

func (s *service) GetFareEstimate(ctx context.Context, req dto.FareEstimateRequest) (*dto.FareEstimateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Validate minimum distance (0.5 km)
	distance := location.HaversineDistance(req.PickupLat, req.PickupLon, req.DropoffLat, req.DropoffLon)
	if distance < 0.5 {
		return nil, response.BadRequest("Minimum trip distance is 0.5 km")
	}

	// Validate maximum distance (100 km)
	if distance > 100 {
		return nil, response.BadRequest("Maximum trip distance is 100 km")
	}

	// Get vehicle type
	vehicleType, err := s.vehiclesRepo.FindByID(ctx, req.VehicleTypeID)
	if err != nil {
		return nil, response.NotFoundError("Vehicle type")
	}

	if !vehicleType.IsActive {
		return nil, response.BadRequest("Vehicle type is not available")
	}

	// Get surge multiplier for pickup location
	surgeMultiplier, err := s.surgeManager.GetSurgeMultiplier(ctx, req.PickupLat, req.PickupLon)
	if err != nil {
		logger.Error("failed to get surge multiplier", "error", err)
		surgeMultiplier = 1.0 // Default to no surge on error
	}

	// Calculate fare estimate
	estimate := s.calculator.CalculateEstimate(
		req.PickupLat, req.PickupLon,
		req.DropoffLat, req.DropoffLon,
		vehicleType,
		surgeMultiplier,
	)

	// Convert to response
	fareResponse := &dto.FareEstimateResponse{
		BaseFare:          estimate.BaseFare,
		DistanceFare:      estimate.DistanceFare,
		DurationFare:      estimate.DurationFare,
		BookingFee:        estimate.BookingFee,
		SurgeMultiplier:   estimate.SurgeMultiplier,
		SubTotal:          estimate.SubTotal,
		SurgeAmount:       estimate.SurgeAmount,
		TotalFare:         estimate.TotalFare,
		EstimatedDistance: estimate.EstimatedDistance,
		EstimatedDuration: estimate.EstimatedDuration,
		VehicleTypeName:   estimate.VehicleTypeName,
		Currency:          "USD",
	}

	// Cache estimate for 1 minute
	cacheKey := fmt.Sprintf("fare:estimate:%s:%f:%f:%f:%f",
		req.VehicleTypeID, req.PickupLat, req.PickupLon, req.DropoffLat, req.DropoffLon)
	cache.SetJSON(ctx, cacheKey, fareResponse, 1*time.Minute)

	logger.Info("fare estimate calculated",
		"vehicleType", vehicleType.Name,
		"distance", estimate.EstimatedDistance,
		"duration", estimate.EstimatedDuration,
		"surge", surgeMultiplier,
		"totalFare", estimate.TotalFare,
	)

	return fareResponse, nil
}

func (s *service) CalculateActualFare(ctx context.Context, req dto.CalculateActualFareRequest) (*dto.FareEstimateResponse, error) {
	if req.ActualDistanceKm <= 0 {
		return nil, response.BadRequest("Invalid distance")
	}
	if req.ActualDurationSec <= 0 {
		return nil, response.BadRequest("Invalid duration")
	}

	// Get vehicle type
	vehicleType, err := s.vehiclesRepo.FindByID(ctx, req.VehicleTypeID)
	if err != nil {
		return nil, response.NotFoundError("Vehicle type")
	}

	// Use provided surge or default to 1.0
	surgeMultiplier := req.SurgeMultiplier
	if surgeMultiplier == 0 {
		surgeMultiplier = 1.0
	}

	// Calculate actual fare
	estimate := s.calculator.CalculateActualFare(
		req.ActualDistanceKm,
		req.ActualDurationSec,
		vehicleType,
		surgeMultiplier,
	)

	// Convert to response
	fareResponse := &dto.FareEstimateResponse{
		BaseFare:          estimate.BaseFare,
		DistanceFare:      estimate.DistanceFare,
		DurationFare:      estimate.DurationFare,
		BookingFee:        estimate.BookingFee,
		SurgeMultiplier:   estimate.SurgeMultiplier,
		SubTotal:          estimate.SubTotal,
		SurgeAmount:       estimate.SurgeAmount,
		TotalFare:         estimate.TotalFare,
		EstimatedDistance: estimate.EstimatedDistance,
		EstimatedDuration: estimate.EstimatedDuration,
		VehicleTypeName:   estimate.VehicleTypeName,
		Currency:          "USD",
	}

	logger.Info("actual fare calculated",
		"vehicleType", vehicleType.Name,
		"distance", req.ActualDistanceKm,
		"duration", req.ActualDurationSec,
		"surge", surgeMultiplier,
		"totalFare", estimate.TotalFare,
	)

	return fareResponse, nil
}

func (s *service) GetSurgeMultiplier(ctx context.Context, lat, lon float64) (float64, error) {
	// Validate coordinates
	if err := location.ValidateCoordinates(lat, lon); err != nil {
		return 1.0, response.BadRequest(err.Error())
	}

	// Try cache first
	geohash := location.Encode(lat, lon, 7) // Precision 7 for ~600m
	cacheKey := fmt.Sprintf("surge:zone:%s", geohash)

	var cached float64
	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		// Parse cached value
		return cached, nil
	}

	// Get from surge manager
	multiplier, err := s.surgeManager.GetSurgeMultiplier(ctx, lat, lon)
	if err != nil {
		logger.Error("failed to get surge multiplier", "error", err)
		return 1.0, nil // Default to no surge on error
	}

	// Cache for 5 minutes
	cache.Set(ctx, cacheKey, fmt.Sprintf("%.2f", multiplier), 5*time.Minute)

	return multiplier, nil
}

func (s *service) GetActiveSurgeZones(ctx context.Context) ([]*dto.SurgeZoneResponse, error) {
	// Try cache first
	cacheKey := "surge:zones:active"
	var cached []*dto.SurgeZoneResponse

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		return cached, nil
	}

	// Get from database
	zones, err := s.repo.GetAllActiveSurgeZones(ctx)
	if err != nil {
		logger.Error("failed to get active surge zones", "error", err)
		return nil, response.InternalServerError("Failed to fetch surge zones", err)
	}

	// Convert to response
	result := make([]*dto.SurgeZoneResponse, len(zones))
	for i, zone := range zones {
		result[i] = &dto.SurgeZoneResponse{
			ID:         zone.ID,
			AreaName:   zone.AreaName,
			Multiplier: zone.Multiplier,
			RadiusKm:   zone.RadiusKm,
			IsActive:   zone.IsActive,
		}
	}

	// Cache for 5 minutes
	cache.SetJSON(ctx, cacheKey, result, 5*time.Minute)

	return result, nil
}

func (s *service) GetFareBreakdown(ctx context.Context, estimate *models.FareEstimate) *dto.FareBreakdownResponse {
	components := []dto.FareComponent{
		{
			Name:   "Base Fare",
			Amount: estimate.BaseFare,
			Type:   "base",
		},
		{
			Name:   fmt.Sprintf("Distance (%.2f km)", estimate.EstimatedDistance),
			Amount: estimate.DistanceFare,
			Type:   "distance",
		},
		{
			Name:   fmt.Sprintf("Duration (%d min)", estimate.EstimatedDuration/60),
			Amount: estimate.DurationFare,
			Type:   "duration",
		},
		{
			Name:   "Booking Fee",
			Amount: estimate.BookingFee,
			Type:   "booking_fee",
		},
	}

	// Add surge if applicable
	if estimate.SurgeMultiplier > 1.0 {
		components = append(components, dto.FareComponent{
			Name:   fmt.Sprintf("Surge (%.1fx)", estimate.SurgeMultiplier),
			Amount: estimate.SurgeAmount,
			Type:   "surge",
		})
	}

	return &dto.FareBreakdownResponse{
		Components: components,
		Total:      estimate.TotalFare,
	}
}
