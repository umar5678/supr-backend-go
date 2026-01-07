package ratings

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
)

type Repository interface {
	Create(ctx context.Context, rating *models.Rating) error
	FindByOrderID(ctx context.Context, orderID string) (*models.Rating, error)
	GetProviderRatings(ctx context.Context, providerID string, limit int) ([]models.Rating, error)
	GetProviderAverageRating(ctx context.Context, providerID string) (float64, error)
	UpdateProviderRating(ctx context.Context, providerID string, newAverage float64) error

	RateDriver(ctx context.Context, rideID string, rating int, comment string) error
	RateRider(ctx context.Context, rideID string, rating int, comment string) error

	// Get ratings
	GetDriverRating(ctx context.Context, driverID string) (float64, error)
	GetRiderRating(ctx context.Context, riderID string) (float64, error)
	GetDriverProfile(ctx context.Context, userID string) (*models.DriverProfile, error)
	GetRiderProfile(ctx context.Context, userID string) (*models.RiderProfile, error)

	// Create profiles if not exists
	CreateRiderProfile(ctx context.Context, profile *models.RiderProfile) error

	// Update profiles
	UpdateDriverRating(ctx context.Context, driverID string) error
	UpdateRiderRating(ctx context.Context, riderID string) error

	// Rating breakdown
	GetDriverRatingBreakdown(ctx context.Context, driverID string) (map[int]int, error)
	GetRiderRatingBreakdown(ctx context.Context, riderID string) (map[int]int, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, rating *models.Rating) error {
	return r.db.WithContext(ctx).Create(rating).Error
}

func (r *repository) FindByOrderID(ctx context.Context, orderID string) (*models.Rating, error) {
	var rating models.Rating
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&rating).Error
	return &rating, err
}

func (r *repository) GetProviderRatings(ctx context.Context, providerID string, limit int) ([]models.Rating, error) {
	var ratings []models.Rating
	err := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID).
		Order("created_at DESC").
		Limit(limit).
		Find(&ratings).Error
	return ratings, err
}

func (r *repository) GetProviderAverageRating(ctx context.Context, providerID string) (float64, error) {
	var avg float64
	err := r.db.WithContext(ctx).
		Model(&models.Rating{}).
		Where("provider_id = ?", providerID).
		Select("COALESCE(AVG(score), 5.0)").
		Scan(&avg).Error
	return avg, err
}

func (r *repository) UpdateProviderRating(ctx context.Context, providerID string, newAverage float64) error {
	return r.db.WithContext(ctx).
		Model(&models.ServiceProvider{}).
		Where("id = ?", providerID).
		Update("rating", newAverage).Error
}

func (r *repository) RateDriver(ctx context.Context, rideID string, rating int, comment string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Where("id = ?", rideID).
		Updates(map[string]interface{}{
			"driver_rating":         rating,
			"driver_rating_comment": comment,
			"driver_rated_at":       now,
		}).Error
}

func (r *repository) RateRider(ctx context.Context, rideID string, rating int, comment string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Where("id = ?", rideID).
		Updates(map[string]interface{}{
			"rider_rating":         rating,
			"rider_rating_comment": comment,
			"rider_rated_at":       now,
		}).Error
}

func (r *repository) GetDriverRating(ctx context.Context, driverID string) (float64, error) {
	var avgRating float64
	err := r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Where("driver_id = ? AND driver_rating IS NOT NULL", driverID).
		Select("COALESCE(AVG(driver_rating), 5.0)").
		Scan(&avgRating).Error
	return avgRating, err
}

func (r *repository) GetRiderRating(ctx context.Context, riderID string) (float64, error) {
	var avgRating float64
	err := r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Where("rider_id = ? AND rider_rating IS NOT NULL", riderID).
		Select("COALESCE(AVG(rider_rating), 5.0)").
		Scan(&avgRating).Error
	return avgRating, err
}

func (r *repository) GetDriverProfile(ctx context.Context, userID string) (*models.DriverProfile, error) {
	var profile models.DriverProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&profile).Error
	return &profile, err
}

func (r *repository) GetRiderProfile(ctx context.Context, userID string) (*models.RiderProfile, error) {
	var profile models.RiderProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&profile).Error
	return &profile, err
}

func (r *repository) CreateRiderProfile(ctx context.Context, profile *models.RiderProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *repository) UpdateDriverRating(ctx context.Context, driverID string) error {
	avgRating, _ := r.GetDriverRating(ctx, driverID)

	return r.db.WithContext(ctx).
		Model(&models.DriverProfile{}).
		Where("user_id = ?", driverID).
		Update("rating", avgRating).Error
}

func (r *repository) UpdateRiderRating(ctx context.Context, riderID string) error {
	avgRating, _ := r.GetRiderRating(ctx, riderID)

	// Get or create rider profile
	var profile models.RiderProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", riderID).First(&profile).Error

	if err == gorm.ErrRecordNotFound {
		// Create new profile
		profile = models.RiderProfile{
			UserID: riderID,
			Rating: avgRating,
		}
		return r.db.WithContext(ctx).Create(&profile).Error
	}

	return r.db.WithContext(ctx).
		Model(&models.RiderProfile{}).
		Where("user_id = ?", riderID).
		Update("rating", avgRating).Error
}

func (r *repository) GetDriverRatingBreakdown(ctx context.Context, driverID string) (map[int]int, error) {
	var results []struct {
		Rating int
		Count  int
	}

	err := r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Select("driver_rating as rating, COUNT(*) as count").
		Where("driver_id = ? AND driver_rating IS NOT NULL", driverID).
		Group("driver_rating").
		Scan(&results).Error

	breakdown := make(map[int]int)
	for i := 1; i <= 5; i++ {
		breakdown[i] = 0
	}

	for _, result := range results {
		breakdown[result.Rating] = result.Count
	}

	return breakdown, err
}

func (r *repository) GetRiderRatingBreakdown(ctx context.Context, riderID string) (map[int]int, error) {
	var results []struct {
		Rating int
		Count  int
	}

	err := r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Select("rider_rating as rating, COUNT(*) as count").
		Where("rider_id = ? AND rider_rating IS NOT NULL", riderID).
		Group("rider_rating").
		Scan(&results).Error

	breakdown := make(map[int]int)
	for i := 1; i <= 5; i++ {
		breakdown[i] = 0
	}

	for _, result := range results {
		breakdown[result.Rating] = result.Count
	}

	return breakdown, err
}
