package riders

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	riderdto "github.com/umar5678/go-backend/internal/modules/riders/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
	"gorm.io/gorm"
)

type Service interface {
	GetProfile(ctx context.Context, userID string) (*riderdto.RiderProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req riderdto.UpdateProfileRequest) (*riderdto.RiderProfileResponse, error)
	GetStats(ctx context.Context, userID string) (*riderdto.RiderStatsResponse, error)

	// Internal methods (used by other modules)
	CreateProfile(ctx context.Context, userID string) (*models.RiderProfile, error)
	IncrementRides(ctx context.Context, userID string) error
	UpdateRating(ctx context.Context, userID string, newRating float64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// GetProfile retrieves rider profile
func (s *service) GetProfile(ctx context.Context, userID string) (*riderdto.RiderProfileResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("rider:profile:%s", userID)
	var cachedProfile models.RiderProfile
	err := cache.GetJSON(ctx, cacheKey, &cachedProfile)
	if err == nil {
		return riderdto.ToRiderProfileResponse(&cachedProfile), nil
	}

	// Get from database
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Rider profile not found")
		}
		return nil, response.InternalServerError("Failed to fetch profile", err)
	}

	// Cache for 5 minutes
	cache.SetJSON(ctx, cacheKey, profile, 5*time.Minute)

	return riderdto.ToRiderProfileResponse(profile), nil
}

// UpdateProfile updates rider profile
func (s *service) UpdateProfile(ctx context.Context, userID string, req riderdto.UpdateProfileRequest) (*riderdto.RiderProfileResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Rider profile not found")
	}

	// Update fields
	if req.HomeAddress != nil {
		profile.HomeAddress = &models.Address{
			Lat:     req.HomeAddress.Lat,
			Lng:     req.HomeAddress.Lng,
			Address: req.HomeAddress.Address,
		}
	}

	if req.WorkAddress != nil {
		profile.WorkAddress = &models.Address{
			Lat:     req.WorkAddress.Lat,
			Lng:     req.WorkAddress.Lng,
			Address: req.WorkAddress.Address,
		}
	}

	if req.PreferredVehicleType != nil {
		profile.PreferredVehicleType = req.PreferredVehicleType
	}

	if err := s.repo.Update(ctx, profile); err != nil {
		logger.Error("failed to update rider profile", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to update profile", err)
	}

	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("rider:profile:%s", userID))

	// Fetch updated profile with relations
	profile, _ = s.repo.FindByUserID(ctx, userID)

	logger.Info("rider profile updated", "userID", userID)

	return riderdto.ToRiderProfileResponse(profile), nil
}

// GetStats retrieves rider statistics
func (s *service) GetStats(ctx context.Context, userID string) (*riderdto.RiderStatsResponse, error) {
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Rider profile not found")
	}

	stats := &riderdto.RiderStatsResponse{
		TotalRides:    profile.TotalRides,
		Rating:        profile.Rating,
		WalletBalance: profile.Wallet.GetAvailableBalance(),
		MemberSince:   profile.CreatedAt.Format("January 2006"),
	}

	return stats, nil
}

// CreateProfile creates a new rider profile (called during signup)
func (s *service) CreateProfile(ctx context.Context, userID string) (*models.RiderProfile, error) {
	// Check if profile already exists
	_, err := s.repo.FindByUserID(ctx, userID)
	if err == nil {
		return nil, response.ConflictError("Rider profile already exists")
	}

	profile := &models.RiderProfile{
		UserID:     userID,
		Rating:     5.0,
		TotalRides: 0,
	}

	if err := s.repo.Create(ctx, profile); err != nil {
		logger.Error("failed to create rider profile", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create profile", err)
	}

	logger.Info("rider profile created", "userID", userID, "profileID", profile.ID)

	return profile, nil
}

// IncrementRides increments total rides (called after ride completion)
func (s *service) IncrementRides(ctx context.Context, userID string) error {
	if err := s.repo.IncrementTotalRides(ctx, userID); err != nil {
		logger.Error("failed to increment rides", "error", err, "userID", userID)
		return response.InternalServerError("Failed to update ride count", err)
	}

	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("rider:profile:%s", userID))

	logger.Info("rider rides incremented", "userID", userID)

	return nil
}

// UpdateRating updates rider rating (called after receiving rating)
func (s *service) UpdateRating(ctx context.Context, userID string, newRating float64) error {
	if newRating < 0 || newRating > 5 {
		return response.BadRequest("Rating must be between 0 and 5")
	}

	if err := s.repo.UpdateRating(ctx, userID, newRating); err != nil {
		logger.Error("failed to update rating", "error", err, "userID", userID)
		return response.InternalServerError("Failed to update rating", err)
	}

	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("rider:profile:%s", userID))

	logger.Info("rider rating updated", "userID", userID, "newRating", newRating)

	return nil
}
