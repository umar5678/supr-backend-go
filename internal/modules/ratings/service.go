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
}

type service struct {
	repo             Repository
	homeServicesRepo homeservices.Repository
}

func NewService(repo Repository, homeServicesRepo homeservices.Repository) Service {
	return &service{
		repo:             repo,
		homeServicesRepo: homeServicesRepo,
	}
}

func (s *service) CreateRating(ctx context.Context, userID string, req dto.CreateRatingRequest) (*dto.RatingResponse, error) {
	// 1. Verify order exists and belongs to user
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

	// 2. Verify order is completed
	if order.Status != "completed" {
		return nil, response.BadRequest("Can only rate completed orders")
	}

	// 3. Verify provider is assigned
	if order.ProviderID == nil {
		return nil, response.BadRequest("No provider assigned to this order")
	}

	// 4. Check if rating already exists
	_, err = s.repo.FindByOrderID(ctx, req.OrderID)
	if err == nil {
		return nil, response.ConflictError("Order already rated")
	}

	// 5. Create rating
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

	// 6. Update provider's average rating
	newAverage, err := s.repo.GetProviderAverageRating(ctx, *order.ProviderID)
	if err == nil {
		s.repo.UpdateProviderRating(ctx, *order.ProviderID, newAverage)
	}

	logger.Info("rating created", "ratingID", rating.ID, "orderID", req.OrderID, "score", req.Score)

	return dto.ToRatingResponse(rating), nil
}
