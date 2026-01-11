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
	FindRideRequestByRideAndDriver(ctx context.Context, rideID, driverID string) (*models.RideRequest, error)
	FindPendingRequestsForDriver(ctx context.Context, driverID string) ([]*models.RideRequest, error)
	FindPendingRequestsForRide(ctx context.Context, rideID string) ([]*models.RideRequest, error)
	UpdateRideRequestStatus(ctx context.Context, requestID, status string, rejectionReason *string) error
	ExpireOldRequests(ctx context.Context) error
	FindActiveRideByDriverID(ctx context.Context, driverID string) (*models.Ride, error)

	// NEW CRITICAL METHODS
	UpdateRideStatusAndDriver(ctx context.Context, rideID, newStatus, expectedStatus string, driverID string) error
	CancelPendingRequestsExcept(ctx context.Context, rideID, acceptedDriverID string) error

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
			surge_multiplier, wallet_hold_id, rider_notes, requested_at, scheduled_at
		) VALUES (
			?, ?, ?, ?,
			ST_GeomFromText(?, 4326), ?, ?, ?,
			ST_GeomFromText(?, 4326), ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?, ?
		)
	`, ride.ID, ride.RiderID, ride.VehicleTypeID, ride.Status,
		pickupPoint, ride.PickupLat, ride.PickupLon, ride.PickupAddress,
		dropoffPoint, ride.DropoffLat, ride.DropoffLon, ride.DropoffAddress,
		ride.EstimatedDistance, ride.EstimatedDuration, ride.EstimatedFare,
		ride.SurgeMultiplier, ride.WalletHoldID, ride.RiderNotes, ride.RequestedAt, ride.ScheduledAt,
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

	if role, ok := filters["role"].(string); ok {
		if role == "rider" {
			query = query.Where("rider_id = ?", userID)
		} else if role == "driver" {
			query = query.Where("driver_id = ?", userID)
		}
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

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

func (r *repository) FindRideRequestByRideAndDriver(ctx context.Context, rideID, driverID string) (*models.RideRequest, error) {
	var request models.RideRequest
	err := r.db.WithContext(ctx).
		Where("ride_id = ? AND driver_id = ?", rideID, driverID).
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

func (r *repository) FindPendingRequestsForRide(ctx context.Context, rideID string) ([]*models.RideRequest, error) {
	var requests []*models.RideRequest
	err := r.db.WithContext(ctx).
		Preload("Driver").
		Preload("Driver.User").
		Preload("Driver.Vehicle").
		Preload("Driver.Vehicle.VehicleType").
		Where("ride_id = ?", rideID).
		Where("status = ?", "pending").
		Order("sent_at ASC").
		Find(&requests).Error
	return requests, err
}

func (r *repository) UpdateRideRequestStatus(ctx context.Context, requestID, status string, rejectionReason *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status != "pending" {
		updates["responded_at"] = time.Now()
	}

	if rejectionReason != nil {
		updates["rejection_reason"] = *rejectionReason
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

func (r *repository) FindActiveRideByDriverID(ctx context.Context, driverID string) (*models.Ride, error) {
	var ride models.Ride
	err := r.db.WithContext(ctx).
		Where("driver_id = ? AND status IN ?", driverID, []string{"accepted", "arrived", "started"}).
		First(&ride).Error
	return &ride, err
}

// ============================================================================
// CRITICAL NEW METHODS FOR RACE CONDITION FIXES
// ============================================================================

// UpdateRideStatusAndDriver atomically updates ride status and assigns driver
// This prevents race conditions where multiple drivers could accept the same ride
func (r *repository) UpdateRideStatusAndDriver(ctx context.Context, rideID, newStatus, expectedStatus string, driverID string) error {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE rides 
		SET status = ?, driver_id = ?, accepted_at = NOW() 
		WHERE id = ? AND status = ?
	`, newStatus, driverID, rideID, expectedStatus)

	if result.Error != nil {
		return result.Error
	}

	// If no rows were affected, the ride was already accepted
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// CancelPendingRequestsExcept cancels all pending ride requests except the accepted one
// This ensures only one driver gets the ride
func (r *repository) CancelPendingRequestsExcept(ctx context.Context, rideID, acceptedDriverID string) error {
	return r.db.WithContext(ctx).
		Model(&models.RideRequest{}).
		Where("ride_id = ?", rideID).
		Where("driver_id != ?", acceptedDriverID).
		Where("status = ?", "pending").
		Updates(map[string]interface{}{
			"status":       "cancelled_by_system",
			"responded_at": time.Now(),
		}).Error
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
