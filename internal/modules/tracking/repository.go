package tracking

// internal/module/tracking/repository .go

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	SaveLocation(ctx context.Context, location *models.DriverLocation) error
	GetDriverLocation(ctx context.Context, driverID string) (*models.DriverLocation, error)
	GetLocationHistory(ctx context.Context, driverID string, from, to time.Time, limit int) ([]*models.DriverLocation, error)
	FindNearbyDrivers(ctx context.Context, lat, lon, radiusKm float64, vehicleTypeID string, limit int) ([]*models.DriverProfile, error)
	BatchSaveLocations(ctx context.Context, locations []*models.DriverLocation) error

	GetDB() *gorm.DB
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) SaveLocation(ctx context.Context, location *models.DriverLocation) error {
	locationStr := fmt.Sprintf("POINT(%f %f)", location.Longitude, location.Latitude)

	return r.db.WithContext(ctx).Exec(`
		INSERT INTO driver_locations_history 
		(driver_id, location, latitude, longitude, heading, speed, accuracy, timestamp, created_at)
		VALUES (?, ST_GeomFromText(?, 4326), ?, ?, ?, ?, ?, ?, NOW())
	`, location.DriverID, locationStr, location.Latitude, location.Longitude,
		location.Heading, location.Speed, location.Accuracy, location.Timestamp).Error
}

func (r *repository) GetDriverLocation(ctx context.Context, driverID string) (*models.DriverLocation, error) {
	var location models.DriverLocation

	err := r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Order("timestamp DESC").
		First(&location).Error

	return &location, err
}

func (r *repository) GetLocationHistory(ctx context.Context, driverID string, from, to time.Time, limit int) ([]*models.DriverLocation, error) {
	var locations []*models.DriverLocation

	query := r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Where("timestamp BETWEEN ? AND ?", from, to).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&locations).Error
	return locations, err
}

func (r *repository) FindNearbyDrivers(ctx context.Context, lat, lon, radiusKm float64, vehicleTypeID string, limit int) ([]*models.DriverProfile, error) {
	var drivers []*models.DriverProfile

	radiusMeters := radiusKm * 1000
	locationStr := fmt.Sprintf("POINT(%f %f)", lon, lat)

	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("Vehicle").
		Preload("Vehicle.VehicleType").
		Where("status = ?", "online").
		Where("is_verified = ?", true).
		Where("current_location IS NOT NULL").
		Where("ST_DWithin(current_location::geography, ST_GeomFromText(?, 4326)::geography, ?)",
			locationStr, radiusMeters)

	if vehicleTypeID != "" {
		query = query.Joins("JOIN vehicles ON vehicles.driver_id = driver_profiles.id").
			Where("vehicles.vehicle_type_id = ?", vehicleTypeID).
			Where("vehicles.is_active = ?", true)
	}

	err := query.
		Select("driver_profiles.*, ST_Distance(current_location::geography, ST_GeomFromText(?, 4326)::geography) as distance", locationStr).
		Order("distance").
		Limit(limit).
		Find(&drivers).Error

	return drivers, err
}

func (r *repository) BatchSaveLocations(ctx context.Context, locations []*models.DriverLocation) error {
	if len(locations) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).CreateInBatches(locations, 100).Error
}
