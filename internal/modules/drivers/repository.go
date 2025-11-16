// internal/modules/drivers/repository.go
package drivers

import (
	"context"
	"fmt"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	// Driver Profile
	CreateDriver(ctx context.Context, driver *models.DriverProfile) error
	FindDriverByID(ctx context.Context, id string) (*models.DriverProfile, error)
	FindDriverByUserID(ctx context.Context, userID string) (*models.DriverProfile, error)
	FindDriverByLicense(ctx context.Context, licenseNumber string) (*models.DriverProfile, error)
	UpdateDriver(ctx context.Context, driver *models.DriverProfile) error
	UpdateDriverStatus(ctx context.Context, driverID, status string) error
	UpdateDriverLocation(ctx context.Context, driverID string, lat, lng float64, heading int) error

	// Vehicle
	CreateVehicle(ctx context.Context, vehicle *models.Vehicle) error
	FindVehicleByDriverID(ctx context.Context, driverID string) (*models.Vehicle, error)
	FindVehicleByPlate(ctx context.Context, licensePlate string) (*models.Vehicle, error)
	UpdateVehicle(ctx context.Context, vehicle *models.Vehicle) error

	// Statistics
	IncrementTrips(ctx context.Context, driverID string) error
	UpdateEarnings(ctx context.Context, driverID string, amount float64) error
	UpdateRating(ctx context.Context, driverID string, newRating float64) error

	// Nearby drivers (for future ride matching)
	FindNearbyDrivers(ctx context.Context, lat, lng, radiusKm float64, vehicleTypeID string) ([]*models.DriverProfile, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Driver Profile Methods

func (r *repository) CreateDriver(ctx context.Context, driver *models.DriverProfile) error {
	return r.db.WithContext(ctx).Create(driver).Error
}

func (r *repository) FindDriverByID(ctx context.Context, id string) (*models.DriverProfile, error) {
	var driver models.DriverProfile
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Vehicle").
		Preload("Vehicle.VehicleType").
		Where("id = ?", id).
		First(&driver).Error
	return &driver, err
}

func (r *repository) FindDriverByUserID(ctx context.Context, userID string) (*models.DriverProfile, error) {
	var driver models.DriverProfile
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Vehicle").
		Preload("Vehicle.VehicleType").
		Where("user_id = ?", userID).
		First(&driver).Error
	return &driver, err
}

func (r *repository) FindDriverByLicense(ctx context.Context, licenseNumber string) (*models.DriverProfile, error) {
	var driver models.DriverProfile
	err := r.db.WithContext(ctx).
		Where("license_number = ?", licenseNumber).
		First(&driver).Error
	return &driver, err
}

func (r *repository) UpdateDriver(ctx context.Context, driver *models.DriverProfile) error {
	return r.db.WithContext(ctx).Save(driver).Error
}

func (r *repository) UpdateDriverStatus(ctx context.Context, driverID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.DriverProfile{}).
		Where("id = ?", driverID).
		Update("status", status).Error
}

func (r *repository) UpdateDriverLocation(ctx context.Context, driverID string, lat, lng float64, heading int) error {
	// Using PostGIS POINT format
	locationStr := fmt.Sprintf("POINT(%f %f)", lng, lat)

	return r.db.WithContext(ctx).
		Model(&models.DriverProfile{}).
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"current_location": gorm.Expr("ST_GeomFromText(?, 4326)", locationStr),
			"heading":          heading,
		}).Error
}

// Vehicle Methods

func (r *repository) CreateVehicle(ctx context.Context, vehicle *models.Vehicle) error {
	return r.db.WithContext(ctx).Create(vehicle).Error
}

func (r *repository) FindVehicleByDriverID(ctx context.Context, driverID string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	err := r.db.WithContext(ctx).
		Preload("VehicleType").
		Where("driver_id = ?", driverID).
		First(&vehicle).Error
	return &vehicle, err
}

func (r *repository) FindVehicleByPlate(ctx context.Context, licensePlate string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	err := r.db.WithContext(ctx).
		Where("license_plate = ?", licensePlate).
		First(&vehicle).Error
	return &vehicle, err
}

func (r *repository) UpdateVehicle(ctx context.Context, vehicle *models.Vehicle) error {
	return r.db.WithContext(ctx).Save(vehicle).Error
}

// Statistics Methods

func (r *repository) IncrementTrips(ctx context.Context, driverID string) error {
	return r.db.WithContext(ctx).
		Model(&models.DriverProfile{}).
		Where("id = ?", driverID).
		UpdateColumn("total_trips", gorm.Expr("total_trips + ?", 1)).Error
}

func (r *repository) UpdateEarnings(ctx context.Context, driverID string, amount float64) error {
	return r.db.WithContext(ctx).
		Model(&models.DriverProfile{}).
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"total_earnings": gorm.Expr("total_earnings + ?", amount),
			"wallet_balance": gorm.Expr("wallet_balance + ?", amount),
		}).Error
}

func (r *repository) UpdateRating(ctx context.Context, driverID string, newRating float64) error {
	return r.db.WithContext(ctx).
		Model(&models.DriverProfile{}).
		Where("id = ?", driverID).
		Update("rating", newRating).Error
}

// Nearby Drivers (using PostGIS)

func (r *repository) FindNearbyDrivers(ctx context.Context, lat, lng, radiusKm float64, vehicleTypeID string) ([]*models.DriverProfile, error) {
	var drivers []*models.DriverProfile

	// Convert km to meters for ST_DWithin
	radiusMeters := radiusKm * 1000

	locationStr := fmt.Sprintf("POINT(%f %f)", lng, lat)

	query := r.db.WithContext(ctx).
		Preload("Vehicle").
		Preload("Vehicle.VehicleType").
		Where("status = ?", "online").
		Where("is_verified = ?", true).
		Where("ST_DWithin(current_location::geography, ST_GeomFromText(?, 4326)::geography, ?)",
			locationStr, radiusMeters)

	if vehicleTypeID != "" {
		query = query.Joins("JOIN vehicles ON vehicles.driver_id = driver_profiles.id").
			Where("vehicles.vehicle_type_id = ?", vehicleTypeID)
	}

	err := query.
		Order(gorm.Expr("ST_Distance(current_location::geography, ST_GeomFromText(?, 4326)::geography)", locationStr)).
		Limit(20).
		Find(&drivers).Error

	return drivers, err
}
