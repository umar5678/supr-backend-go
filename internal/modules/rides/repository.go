package rides

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	// Ride CRUD
	CreateRide(ctx context.Context, ride *models.Ride) error
	FindRideByID(ctx context.Context, id string) (*models.Ride, error)
	UpdateRide(ctx context.Context, ride *models.Ride) error
	UpdateRideStatus(ctx context.Context, rideID, status string) error
	ListRides(ctx context.Context, userID string, filters map[string]interface{}, page, limit int) ([]*models.Ride, int64, error)

	// Ride Requests
	CreateRideRequest(ctx context.Context, request *models.RideRequest) error
	FindRideRequestByID(ctx context.Context, id string) (*models.RideRequest, error)
	FindPendingRequestsForDriver(ctx context.Context, driverID string) ([]*models.RideRequest, error)
	UpdateRideRequestStatus(ctx context.Context, requestID, status string) error
	ExpireOldRequests(ctx context.Context) error

	// Statistics
	GetRiderStats(ctx context.Context, riderID string) (totalRides int, totalSpent float64, err error)
	GetDriverStats(ctx context.Context, driverID string) (totalTrips int, totalEarnings float64, err error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Ride CRUD

func (r *repository) CreateRide(ctx context.Context, ride *models.Ride) error {
	pickupPoint := fmt.Sprintf("POINT(%f %f)", ride.PickupLon, ride.PickupLat)
	dropoffPoint := fmt.Sprintf("POINT(%f %f)", ride.DropoffLon, ride.DropoffLat)

	return r.db.WithContext(ctx).Exec(`
		INSERT INTO rides (
			id, rider_id, vehicle_type_id, status,
			pickup_location, pickup_lat, pickup_lon, pickup_address,
			dropoff_location, dropoff_lat, dropoff_lon, dropoff_address,
			estimated_distance, estimated_duration, estimated_fare,
			surge_multiplier, wallet_hold_id, rider_notes, requested_at
		) VALUES (
			?, ?, ?, ?,
			ST_GeomFromText(?, 4326), ?, ?, ?,
			ST_GeomFromText(?, 4326), ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?
		)
	`, ride.ID, ride.RiderID, ride.VehicleTypeID, ride.Status,
		pickupPoint, ride.PickupLat, ride.PickupLon, ride.PickupAddress,
		dropoffPoint, ride.DropoffLat, ride.DropoffLon, ride.DropoffAddress,
		ride.EstimatedDistance, ride.EstimatedDuration, ride.EstimatedFare,
		ride.SurgeMultiplier, ride.WalletHoldID, ride.RiderNotes, ride.RequestedAt,
	).Error
}

func (r *repository) FindRideByID(ctx context.Context, id string) (*models.Ride, error) {
	var ride models.Ride
	err := r.db.WithContext(ctx).
		Preload("Rider").
		Preload("Driver").
		Preload("VehicleType").
		Where("id = ?", id).
		First(&ride).Error
	return &ride, err
}

func (r *repository) UpdateRide(ctx context.Context, ride *models.Ride) error {
	return r.db.WithContext(ctx).Save(ride).Error
}

func (r *repository) UpdateRideStatus(ctx context.Context, rideID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// Set timestamps based on status
	now := time.Now()
	switch status {
	case "accepted":
		updates["accepted_at"] = now
	case "arrived":
		updates["arrived_at"] = now
	case "started":
		updates["started_at"] = now
	case "completed":
		updates["completed_at"] = now
	case "cancelled":
		updates["cancelled_at"] = now
	}

	return r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Where("id = ?", rideID).
		Updates(updates).Error
}

func (r *repository) ListRides(ctx context.Context, userID string, filters map[string]interface{}, page, limit int) ([]*models.Ride, int64, error) {
	var rides []*models.Ride
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Ride{})

	// Filter by user (rider or driver)
	if role, ok := filters["role"].(string); ok {
		if role == "rider" {
			query = query.Where("rider_id = ?", userID)
		} else if role == "driver" {
			query = query.Where("driver_id = ?", userID)
		}
	}

	// Filter by status
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	query.Count(&total)

	// Paginate
	offset := (page - 1) * limit
	err := query.
		Preload("Rider").
		Preload("Driver").
		Preload("VehicleType").
		Order("requested_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rides).Error

	return rides, total, err
}

// Ride Requests

func (r *repository) CreateRideRequest(ctx context.Context, request *models.RideRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *repository) FindRideRequestByID(ctx context.Context, id string) (*models.RideRequest, error) {
	var request models.RideRequest
	err := r.db.WithContext(ctx).
		Preload("Ride").
		Preload("Driver").
		Where("id = ?", id).
		First(&request).Error
	return &request, err
}

func (r *repository) FindPendingRequestsForDriver(ctx context.Context, driverID string) ([]*models.RideRequest, error) {
	var requests []*models.RideRequest
	err := r.db.WithContext(ctx).
		Preload("Ride").
		Where("driver_id = ?", driverID).
		Where("status = ?", "pending").
		Where("expires_at > ?", time.Now()).
		Order("sent_at ASC").
		Find(&requests).Error
	return requests, err
}

func (r *repository) UpdateRideRequestStatus(ctx context.Context, requestID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status != "pending" {
		updates["responded_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&models.RideRequest{}).
		Where("id = ?", requestID).
		Updates(updates).Error
}

func (r *repository) ExpireOldRequests(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&models.RideRequest{}).
		Where("status = ?", "pending").
		Where("expires_at < ?", time.Now()).
		Update("status", "expired").Error
}

// Statistics

func (r *repository) GetRiderStats(ctx context.Context, riderID string) (totalRides int, totalSpent float64, err error) {
	var stats struct {
		TotalRides int
		TotalSpent float64
	}

	err = r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Select("COUNT(*) as total_rides, COALESCE(SUM(actual_fare), 0) as total_spent").
		Where("rider_id = ?", riderID).
		Where("status = ?", "completed").
		Scan(&stats).Error

	return stats.TotalRides, stats.TotalSpent, err
}

func (r *repository) GetDriverStats(ctx context.Context, driverID string) (totalTrips int, totalEarnings float64, err error) {
	var stats struct {
		TotalTrips    int
		TotalEarnings float64
	}

	err = r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Select("COUNT(*) as total_trips, COALESCE(SUM(actual_fare * 0.80), 0) as total_earnings").
		Where("driver_id = ?", driverID).
		Where("status = ?", "completed").
		Scan(&stats).Error

	return stats.TotalTrips, stats.TotalEarnings, err
}
