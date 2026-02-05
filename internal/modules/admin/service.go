package admin

// now

import (
	"context"
	"strconv"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/serviceproviders"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	ListUsers(ctx context.Context, role, status, page, limit string) (map[string]interface{}, error)
	ApproveServiceProvider(ctx context.Context, providerID string) error
	SuspendUser(ctx context.Context, userID, reason string) error
	UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error
	GetDashboardStats(ctx context.Context) (map[string]interface{}, error)
}

type service struct {
	repo   Repository
	spRepo serviceproviders.Repository
}

func NewService(repo Repository, spRepo serviceproviders.Repository) Service {
	return &service{
		repo:   repo,
		spRepo: spRepo,
	}
}

func (s *service) ListUsers(ctx context.Context, role, status, page, limit string) (map[string]interface{}, error) {
	filters := make(map[string]interface{})
	if role != "" {
		filters["role"] = role
	}
	if status != "" {
		filters["status"] = status
	}

	pageInt, _ := strconv.Atoi(page)
	if pageInt < 1 {
		pageInt = 1
	}

	limitInt, _ := strconv.Atoi(limit)
	if limitInt < 1 || limitInt > 100 {
		limitInt = 20
	}

	users, total, err := s.repo.ListUsers(ctx, filters, pageInt, limitInt)
	if err != nil {
		return nil, response.InternalServerError("Failed to fetch users", err)
	}

	return map[string]interface{}{
		"users": users,
		"total": total,
		"page":  pageInt,
		"limit": limitInt,
	}, nil
}

func (s *service) ApproveServiceProvider(ctx context.Context, providerID string) error {

	profile, err := s.spRepo.FindByID(ctx, providerID)
	if err != nil {
		return response.NotFoundError("Service provider")
	}

	if err := s.repo.UpdateUserStatus(ctx, profile.UserID, models.StatusActive); err != nil {
		return response.InternalServerError("Failed to update user status", err)
	}

	if err := s.spRepo.UpdateStatus(ctx, providerID, models.SPStatusActive); err != nil {
		return response.InternalServerError("Failed to update provider status", err)
	}

	logger.Info("service provider approved", "providerID", providerID, "userID", profile.UserID)
	return nil
}

func (s *service) SuspendUser(ctx context.Context, userID, reason string) error {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return response.NotFoundError("User")
	}

	if err := s.repo.UpdateUserStatus(ctx, userID, models.StatusSuspended); err != nil {
		return response.InternalServerError("Failed to suspend user", err)
	}

	logger.Info("user suspended", "userID", userID, "reason", reason, user)
	return nil
}

func (s *service) UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error {
	if err := s.repo.UpdateUserStatus(ctx, userID, status); err != nil {
		return response.InternalServerError("Failed to update user status", err)
	}

	logger.Info("user status updated", "userID", userID, "status", status)
	return nil
}

func (s *service) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := s.repo.GetDashboardStats(ctx)
	if err != nil {
		return nil, response.InternalServerError("Failed to fetch stats", err)
	}
	return stats, nil
}
