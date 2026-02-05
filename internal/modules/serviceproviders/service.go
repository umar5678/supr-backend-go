package serviceproviders

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	CreateProfile(ctx context.Context, userID, serviceCategory string) (*models.ServiceProviderProfile, error)
	GetProfile(ctx context.Context, userID string) (*models.ServiceProviderProfile, error)
	UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) (*models.ServiceProviderProfile, error)
	List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.ServiceProviderProfile, int64, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateProfile(ctx context.Context, userID, serviceCategory string) (*models.ServiceProviderProfile, error) {
	profile := &models.ServiceProviderProfile{
		UserID:          userID,
		ServiceCategory: serviceCategory,
		Status:          models.SPStatusPendingApproval,
		IsVerified:      false,
		Rating:          0.0,
		TotalReviews:    0,
		CompletedJobs:   0,
		IsAvailable:     true,
		Currency:        "INR",
	}

	if err := s.repo.CreateProfile(ctx, profile); err != nil {
		logger.Error("failed to create service provider profile", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create profile", err)
	}

	return profile, nil
}

func (s *service) GetProfile(ctx context.Context, userID string) (*models.ServiceProviderProfile, error) {
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Service provider profile")
	}
	return profile, nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) (*models.ServiceProviderProfile, error) {
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Service provider profile")
	}

	if businessName, ok := updates["businessName"].(string); ok {
		profile.BusinessName = &businessName
	}
	if description, ok := updates["description"].(string); ok {
		profile.Description = &description
	}
	if hourlyRate, ok := updates["hourlyRate"].(float64); ok {
		profile.HourlyRate = &hourlyRate
	}
	if isAvailable, ok := updates["isAvailable"].(bool); ok {
		profile.IsAvailable = isAvailable
	}

	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, response.InternalServerError("Failed to update profile", err)
	}

	return profile, nil
}

func (s *service) List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.ServiceProviderProfile, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.repo.List(ctx, filters, page, limit)
}
