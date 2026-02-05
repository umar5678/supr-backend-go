package ratings

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices"
	"github.com/umar5678/go-backend/internal/modules/ratings/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	CreateRating(ctx context.Context, userID string, req dto.CreateRatingRequest) (*dto.RatingResponse, error)
	RateDriver(ctx context.Context, riderID string, req dto.RateDriverRequest) error
	RateRider(ctx context.Context, driverID string, req dto.RateRiderRequest) error
	GetDriverRatingStats(ctx context.Context, driverID string) (*dto.RatingStatsResponse, error)
	GetRiderRatingStats(ctx context.Context, riderID string) (*dto.RatingStatsResponse, error)
	GetDriverRatingBreakdown(ctx context.Context, driverID string) (*dto.RatingBreakdownResponse, error)
	GetRiderRatingBreakdown(ctx context.Context, riderID string) (*dto.RatingBreakdownResponse, error)
}

type service struct {
	repo             Repository
	db               *gorm.DB
	homeServicesRepo homeservices.Repository
}

func NewService(repo Repository, db *gorm.DB, homeServicesRepo homeservices.Repository) Service {
	return &service{
		repo:             repo,
		homeServicesRepo: homeServicesRepo,
		db:               db,
	}
}

func (s *service) CreateRating(ctx context.Context, userID string, req dto.CreateRatingRequest) (*dto.RatingResponse, error) {
	order, err := s.homeServicesRepo.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to fetch order", err)
	}

	if order.UserID != userID {
		return nil, response.ForbiddenError("You don't have access to this order")
	}

	if order.Status != "completed" {
		return nil, response.BadRequest("Can only rate completed orders")
	}

	if order.ProviderID == nil {
		return nil, response.BadRequest("No provider assigned to this order")
	}

	_, err = s.repo.FindByOrderID(ctx, req.OrderID)
	if err == nil {
		return nil, response.ConflictError("Order already rated")
	}

	rating := &models.Rating{
		ID:         uuid.New().String(),
		OrderID:    req.OrderID,
		UserID:     userID,
		ProviderID: *order.ProviderID,
		Score:      req.Score,
		Comment:    req.Comment,
	}

	if err := s.repo.Create(ctx, rating); err != nil {
		logger.Error("failed to create rating", "error", err)
		return nil, response.InternalServerError("Failed to create rating", err)
	}

	newAverage, err := s.repo.GetProviderAverageRating(ctx, *order.ProviderID)
	if err == nil {
		s.repo.UpdateProviderRating(ctx, *order.ProviderID, newAverage)
	}

	logger.Info("rating created", "ratingID", rating.ID, "orderID", req.OrderID, "score", req.Score)

	return dto.ToRatingResponse(rating), nil
}

func (s *service) RateDriver(ctx context.Context, riderID string, req dto.RateDriverRequest) error {
	if err := req.Validate(); err != nil {
		return response.BadRequest(err.Error())
	}

	var ride models.Ride
	if err := s.db.WithContext(ctx).Where("id = ?", req.RideID).First(&ride).Error; err != nil {
		return response.NotFoundError("Ride")
	}

	if ride.RiderID != riderID {
		return response.ForbiddenError("You can only rate your own rides")
	}

	if ride.Status != "completed" {
		return response.BadRequest("Can only rate completed rides")
	}

	if ride.DriverRating != nil {
		return response.BadRequest("You have already rated this driver")
	}

	if ride.DriverID == nil {
		return response.BadRequest("No driver assigned to this ride")
	}

	if err := s.repo.RateDriver(ctx, req.RideID, req.Rating, req.Comment); err != nil {
		logger.Error("failed to rate driver", "error", err, "rideID", req.RideID)
		return response.InternalServerError("Failed to rate driver", err)
	}

	go s.repo.UpdateDriverRating(context.Background(), *ride.DriverID)

	logger.Info("driver rated",
		"rideID", req.RideID,
		"driverID", *ride.DriverID,
		"rating", req.Rating,
	)

	return nil
}

func (s *service) RateRider(ctx context.Context, driverID string, req dto.RateRiderRequest) error {
	if err := req.Validate(); err != nil {
		return response.BadRequest(err.Error())
	}

	var ride models.Ride
	if err := s.db.WithContext(ctx).Where("id = ?", req.RideID).First(&ride).Error; err != nil {
		return response.NotFoundError("Ride")
	}

	if ride.DriverID == nil || *ride.DriverID != driverID {
		return response.ForbiddenError("You can only rate your own rides")
	}

	if ride.Status != "completed" {
		return response.BadRequest("Can only rate completed rides")
	}

	if ride.RiderRating != nil {
		return response.BadRequest("You have already rated this rider")
	}

	if err := s.repo.RateRider(ctx, req.RideID, req.Rating, req.Comment); err != nil {
		logger.Error("failed to rate rider", "error", err, "rideID", req.RideID)
		return response.InternalServerError("Failed to rate rider", err)
	}

	go s.repo.UpdateRiderRating(context.Background(), ride.RiderID)

	logger.Info("rider rated",
		"rideID", req.RideID,
		"riderID", ride.RiderID,
		"rating", req.Rating,
	)

	return nil
}

func (s *service) GetDriverRatingStats(ctx context.Context, driverID string) (*dto.RatingStatsResponse, error) {
	profile, err := s.repo.GetDriverProfile(ctx, driverID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	return dto.ToDriverRatingStats(profile), nil
}

func (s *service) GetRiderRatingStats(ctx context.Context, riderID string) (*dto.RatingStatsResponse, error) {
	profile, err := s.repo.GetRiderProfile(ctx, riderID)
	if err != nil {
		rating, _ := s.repo.GetRiderRating(ctx, riderID)
		profile = &models.RiderProfile{
			UserID: riderID,
			Rating: rating,
		}
		s.repo.CreateRiderProfile(ctx, profile)
	}

	return dto.ToRiderRatingStats(profile), nil
}

func (s *service) GetDriverRatingBreakdown(ctx context.Context, driverID string) (*dto.RatingBreakdownResponse, error) {
	breakdown, err := s.repo.GetDriverRatingBreakdown(ctx, driverID)
	if err != nil {
		return nil, response.InternalServerError("Failed to get rating breakdown", err)
	}

	totalRatings := 0
	totalScore := 0
	for rating, count := range breakdown {
		totalRatings += count
		totalScore += rating * count
	}

	avgRating := 0.0
	if totalRatings > 0 {
		avgRating = float64(totalScore) / float64(totalRatings)
	}

	return &dto.RatingBreakdownResponse{
		FiveStars:     breakdown[5],
		FourStars:     breakdown[4],
		ThreeStars:    breakdown[3],
		TwoStars:      breakdown[2],
		OneStar:       breakdown[1],
		TotalRatings:  totalRatings,
		AverageRating: avgRating,
	}, nil
}

func (s *service) GetRiderRatingBreakdown(ctx context.Context, riderID string) (*dto.RatingBreakdownResponse, error) {
	breakdown, err := s.repo.GetRiderRatingBreakdown(ctx, riderID)
	if err != nil {
		return nil, response.InternalServerError("Failed to get rating breakdown", err)
	}

	totalRatings := 0
	totalScore := 0
	for rating, count := range breakdown {
		totalRatings += count
		totalScore += rating * count
	}

	avgRating := 0.0
	if totalRatings > 0 {
		avgRating = float64(totalScore) / float64(totalRatings)
	}

	return &dto.RatingBreakdownResponse{
		FiveStars:     breakdown[5],
		FourStars:     breakdown[4],
		ThreeStars:    breakdown[3],
		TwoStars:      breakdown[2],
		OneStar:       breakdown[1],
		TotalRatings:  totalRatings,
		AverageRating: avgRating,
	}, nil
}
