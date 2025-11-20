// internal/modules/drivers/service.go
package drivers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	driverdto "github.com/umar5678/go-backend/internal/modules/drivers/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/helpers"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
	"gorm.io/gorm"
)

type Service interface {
	RegisterDriver(ctx context.Context, userID string, req driverdto.RegisterDriverRequest) (*driverdto.DriverProfileResponse, error)
	GetProfile(ctx context.Context, userID string) (*driverdto.DriverProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req driverdto.UpdateDriverProfileRequest) (*driverdto.DriverProfileResponse, error)
	UpdateVehicle(ctx context.Context, userID string, req driverdto.UpdateVehicleRequest) (*driverdto.VehicleResponse, error)
	UpdateStatus(ctx context.Context, userID string, req driverdto.UpdateStatusRequest) (*driverdto.DriverProfileResponse, error)
	UpdateLocation(ctx context.Context, userID string, req driverdto.UpdateLocationRequest) error
	GetWallet(ctx context.Context, userID string) (*driverdto.WalletResponse, error)
	GetDashboard(ctx context.Context, userID string) (*driverdto.DriverDashboardResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) RegisterDriver(ctx context.Context, userID string, req driverdto.RegisterDriverRequest) (*driverdto.DriverProfileResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if user is already a driver
	_, err := s.repo.FindDriverByUserID(ctx, userID)
	if err == nil {
		return nil, response.BadRequest("User is already registered as a driver")
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("failed to check existing driver", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to check driver status", err)
	}

	// Check if license number already exists
	_, err = s.repo.FindDriverByLicense(ctx, req.LicenseNumber)
	if err == nil {
		return nil, response.BadRequest("License number is already registered")
	}

	// Check if license plate already exists
	_, err = s.repo.FindVehicleByPlate(ctx, req.Vehicle.LicensePlate)
	if err == nil {
		return nil, response.BadRequest("License plate is already registered")
	}

	// Create driver profile
	driver := &models.DriverProfile{
		UserID:        userID,
		LicenseNumber: req.LicenseNumber,
		Status:        "online",
		Rating:        5.0,
		IsVerified:    true, // Auto-approve (no document verification needed)
	}

	if err := s.repo.CreateDriver(ctx, driver); err != nil {
		logger.Error("failed to create driver profile", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create driver profile", err)
	}

	// Create vehicle
	vehicle := &models.Vehicle{
		DriverID:      driver.ID,
		VehicleTypeID: req.Vehicle.VehicleTypeID,
		Make:          req.Vehicle.Make,
		Model:         req.Vehicle.Model,
		Year:          req.Vehicle.Year,
		Color:         req.Vehicle.Color,
		LicensePlate:  req.Vehicle.LicensePlate,
		Capacity:      req.Vehicle.Capacity,
		IsActive:      true,
	}

	if err := s.repo.CreateVehicle(ctx, vehicle); err != nil {
		logger.Error("failed to create vehicle", "error", err, "driverID", driver.ID)
		return nil, response.InternalServerError("Failed to register vehicle", err)
	}

	// Fetch complete profile with relations
	driver, _ = s.repo.FindDriverByID(ctx, driver.ID)

	logger.Info("driver registered successfully",
		"driverID", driver.ID,
		"userID", userID,
		"vehicleType", req.Vehicle.VehicleTypeID,
	)

	return driverdto.ToDriverProfileResponse(driver), nil
}

func (s *service) GetProfile(ctx context.Context, userID string) (*driverdto.DriverProfileResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("driver:profile:%s", userID)
	var cached models.DriverProfile

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		logger.Debug("driver profile cache hit", "userID", userID)
		return driverdto.ToDriverProfileResponse(&cached), nil
	}

	// Get from database
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("Driver profile")
		}
		logger.Error("failed to fetch driver profile", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to fetch profile", err)
	}

	// Cache for 5 minutes
	cache.SetJSON(ctx, cacheKey, driver, 5*time.Minute)

	return driverdto.ToDriverProfileResponse(driver), nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, req driverdto.UpdateDriverProfileRequest) (*driverdto.DriverProfileResponse, error) {
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	// Update fields
	if req.LicenseNumber != nil {
		// Check if new license number is unique
		existing, err := s.repo.FindDriverByLicense(ctx, *req.LicenseNumber)
		if err == nil && existing.ID != driver.ID {
			return nil, response.BadRequest("License number is already in use")
		}
		driver.LicenseNumber = *req.LicenseNumber
	}

	if err := s.repo.UpdateDriver(ctx, driver); err != nil {
		logger.Error("failed to update driver profile", "error", err, "driverID", driver.ID)
		return nil, response.InternalServerError("Failed to update profile", err)
	}

	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("driver:profile:%s", userID))

	// Fetch updated profile
	driver, _ = s.repo.FindDriverByID(ctx, driver.ID)

	logger.Info("driver profile updated", "driverID", driver.ID, "userID", userID)

	return driverdto.ToDriverProfileResponse(driver), nil
}

func (s *service) UpdateVehicle(ctx context.Context, userID string, req driverdto.UpdateVehicleRequest) (*driverdto.VehicleResponse, error) {
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	vehicle, err := s.repo.FindVehicleByDriverID(ctx, driver.ID)
	if err != nil {
		return nil, response.NotFoundError("Vehicle")
	}

	// Update fields
	if req.VehicleTypeID != nil {
		vehicle.VehicleTypeID = *req.VehicleTypeID
	}
	if req.Make != nil {
		vehicle.Make = *req.Make
	}
	if req.Model != nil {
		vehicle.Model = *req.Model
	}
	if req.Year != nil {
		vehicle.Year = *req.Year
	}
	if req.Color != nil {
		vehicle.Color = *req.Color
	}
	if req.LicensePlate != nil {
		// Check if plate is unique
		existing, err := s.repo.FindVehicleByPlate(ctx, *req.LicensePlate)
		if err == nil && existing.ID != vehicle.ID {
			return nil, response.BadRequest("License plate is already in use")
		}
		vehicle.LicensePlate = *req.LicensePlate
	}
	if req.Capacity != nil {
		vehicle.Capacity = *req.Capacity
	}
	if req.IsActive != nil {
		vehicle.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateVehicle(ctx, vehicle); err != nil {
		logger.Error("failed to update vehicle", "error", err, "vehicleID", vehicle.ID)
		return nil, response.InternalServerError("Failed to update vehicle", err)
	}

	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("driver:profile:%s", userID))

	// Fetch updated vehicle
	vehicle, _ = s.repo.FindVehicleByDriverID(ctx, driver.ID)

	logger.Info("vehicle updated", "vehicleID", vehicle.ID, "driverID", driver.ID)

	return driverdto.ToVehicleResponse(vehicle), nil
}

func (s *service) UpdateStatus(ctx context.Context, userID string, req driverdto.UpdateStatusRequest) (*driverdto.DriverProfileResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	// Check if driver can go online
	if req.Status == "online" {
		if !driver.IsVerified {
			return nil, response.BadRequest("Driver is not verified yet")
		}

		// Check if vehicle exists and is active
		vehicle, err := s.repo.FindVehicleByDriverID(ctx, driver.ID)
		if err != nil {
			return nil, response.BadRequest("Vehicle information is required to go online")
		}
		if !vehicle.IsActive {
			return nil, response.BadRequest("Vehicle is not active")
		}
	}

	oldStatus := driver.Status

	if err := s.repo.UpdateDriverStatus(ctx, driver.ID, req.Status); err != nil {
		logger.Error("failed to update driver status", "error", err, "driverID", driver.ID)
		return nil, response.InternalServerError("Failed to update status", err)
	}

	// Update online status in Redis
	onlineKey := fmt.Sprintf("driver:online:%s", driver.ID)
	if req.Status == "online" {
		cache.Set(ctx, onlineKey, "true", 5*time.Minute) // TTL for heartbeat

		// Add to online drivers set
		cache.SessionClient.SAdd(ctx, "drivers:online", driver.ID)
	} else {
		cache.Delete(ctx, onlineKey)
		cache.SessionClient.SRem(ctx, "drivers:online", driver.ID)
	}

	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("driver:profile:%s", userID))

	// Broadcast status change via WebSocket
	go func() {
		helpers.BroadcastNotification(map[string]interface{}{
			"type":      "driver_status_changed",
			"driverId":  driver.ID,
			"status":    req.Status,
			"oldStatus": oldStatus,
		})
	}()

	// Fetch updated profile
	driver, _ = s.repo.FindDriverByID(ctx, driver.ID)

	logger.Info("driver status updated",
		"driverID", driver.ID,
		"oldStatus", oldStatus,
		"newStatus", req.Status,
	)

	return driverdto.ToDriverProfileResponse(driver), nil
}

func (s *service) UpdateLocation(ctx context.Context, userID string, req driverdto.UpdateLocationRequest) error {
	if err := req.Validate(); err != nil {
		return response.BadRequest(err.Error())
	}

	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return response.NotFoundError("Driver profile")
	}

	// Only allow location updates if driver is online
	// if driver.Status != "online" && driver.Status != "busy" && driver.Status != "on_trip" {
	// 	return response.BadRequest("Driver must be online to update location")
	// }
	// log.Printf()
	if driver.Status != "online" && driver.Status != "busy" {
		return response.BadRequest("Driver must be online to update location ok")
	}

	// Update location in database
	if err := s.repo.UpdateDriverLocation(ctx, driver.ID, req.Latitude, req.Longitude, req.Heading); err != nil {
		logger.Error("failed to update driver location", "error", err, "driverID", driver.ID)
		return response.InternalServerError("Failed to update location", err)
	}

	// Update location in Redis (for faster access)
	locationKey := fmt.Sprintf("driver:location:%s", driver.ID)
	locationData := map[string]interface{}{
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"heading":   req.Heading,
		"timestamp": time.Now().Unix(),
	}
	cache.SetJSON(ctx, locationKey, locationData, 30*time.Second)

	// Refresh online status TTL
	onlineKey := fmt.Sprintf("driver:online:%s", driver.ID)
	cache.Set(ctx, onlineKey, "true", 5*time.Minute)

	logger.Debug("driver location updated",
		"driverID", driver.ID,
		"lat", req.Latitude,
		"lng", req.Longitude,
	)

	return nil
}

func (s *service) GetWallet(ctx context.Context, userID string) (*driverdto.WalletResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("driver:wallet:%s", userID)
	var cached driverdto.WalletResponse

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		return &cached, nil
	}

	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	walletResp := &driverdto.WalletResponse{
		Balance:       driver.WalletBalance,
		TotalEarnings: driver.TotalEarnings,
		PendingAmount: 0, // TODO: Calculate from pending rides
	}

	// Cache for 1 minute
	cache.SetJSON(ctx, cacheKey, walletResp, 1*time.Minute)

	return walletResp, nil
}

func (s *service) GetDashboard(ctx context.Context, userID string) (*driverdto.DriverDashboardResponse, error) {
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	// TODO: Calculate today's trips and earnings from rides table
	// For now, returning mock data
	dashboard := &driverdto.DriverDashboardResponse{
		TodayTrips:     0, // TODO: Query from rides
		TodayEarnings:  0, // TODO: Query from rides
		WeekEarnings:   0, // TODO: Query from rides
		Rating:         driver.Rating,
		AcceptanceRate: driver.AcceptanceRate,
		CompletionRate: 100 - driver.CancellationRate,
		TotalTrips:     driver.TotalTrips,
		WalletBalance:  driver.WalletBalance,
		Status:         driver.Status,
		Profile:        driverdto.ToDriverProfileResponse(driver),
	}

	return dashboard, nil
}
