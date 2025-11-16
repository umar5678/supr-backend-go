package vehicles

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/modules/vehicles/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	GetAllVehicleTypes(ctx context.Context) ([]*dto.VehicleTypeResponse, error)
	GetActiveVehicleTypes(ctx context.Context) ([]*dto.VehicleTypeResponse, error)
	GetVehicleTypeByID(ctx context.Context, id string) (*dto.VehicleTypeResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAllVehicleTypes(ctx context.Context) ([]*dto.VehicleTypeResponse, error) {
	// Try cache first
	cacheKey := "vehicle:types:all"
	var cached []*dto.VehicleTypeResponse

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil && len(cached) > 0 {
		logger.Debug("vehicle types cache hit", "key", cacheKey)
		return cached, nil
	}

	// Get from database
	vehicleTypes, err := s.repo.FindAll(ctx)
	if err != nil {
		logger.Error("failed to fetch vehicle types", "error", err)
		return nil, response.InternalServerError("Failed to fetch vehicle types", err)
	}

	// Convert to response
	result := make([]*dto.VehicleTypeResponse, len(vehicleTypes))
	for i, vt := range vehicleTypes {
		result[i] = dto.ToVehicleTypeResponse(vt)
	}

	// Cache for 10 minutes
	cache.SetJSON(ctx, cacheKey, result, 10*time.Minute)

	return result, nil
}

func (s *service) GetActiveVehicleTypes(ctx context.Context) ([]*dto.VehicleTypeResponse, error) {
	// Try cache first
	cacheKey := "vehicle:types:active"
	var cached []*dto.VehicleTypeResponse

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil && len(cached) > 0 {
		logger.Debug("active vehicle types cache hit", "key", cacheKey)
		return cached, nil
	}

	// Get from database
	vehicleTypes, err := s.repo.FindActive(ctx)
	if err != nil {
		logger.Error("failed to fetch active vehicle types", "error", err)
		return nil, response.InternalServerError("Failed to fetch vehicle types", err)
	}

	// Convert to response
	result := make([]*dto.VehicleTypeResponse, len(vehicleTypes))
	for i, vt := range vehicleTypes {
		result[i] = dto.ToVehicleTypeResponse(vt)
	}

	// Cache for 10 minutes
	cache.SetJSON(ctx, cacheKey, result, 10*time.Minute)

	logger.Info("active vehicle types retrieved", "count", len(result))

	return result, nil
}

func (s *service) GetVehicleTypeByID(ctx context.Context, id string) (*dto.VehicleTypeResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("vehicle:type:%s", id)
	var cached dto.VehicleTypeResponse

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		return &cached, nil
	}

	// Get from database
	vehicleType, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, response.NotFoundError("Vehicle type")
	}

	result := dto.ToVehicleTypeResponse(vehicleType)

	// Cache for 10 minutes
	cache.SetJSON(ctx, cacheKey, result, 10*time.Minute)

	return result, nil
}
