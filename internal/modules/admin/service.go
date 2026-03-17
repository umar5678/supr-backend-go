package admin

import (
	"context"
	"strconv"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/drivers"
	"github.com/umar5678/go-backend/internal/modules/notifications"
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
	ListDriverProfiles(ctx context.Context, filters map[string]interface{}, page, limit int) (map[string]interface{}, error)
	ListServiceProviderProfiles(ctx context.Context, filters map[string]interface{}, page, limit int) (map[string]interface{}, error)
}

type service struct {
	repo          Repository
	spRepo        serviceproviders.Repository
	drvRepo       drivers.Repository
	eventProducer notifications.EventProducer
}

func NewService(repo Repository, spRepo serviceproviders.Repository, drvRepo drivers.Repository) Service {
	return NewServiceWithNotifications(repo, spRepo, drvRepo, nil)
}

func NewServiceWithNotifications(repo Repository, spRepo serviceproviders.Repository, drvRepo drivers.Repository, eventProducer notifications.EventProducer) Service {
	return &service{
		repo:          repo,
		spRepo:        spRepo,
		drvRepo:       drvRepo,
		eventProducer: eventProducer,
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

	s.publishAdminEvent(ctx, notifications.EventUserSuspended, map[string]interface{}{
		"user_id":   userID,
		"reason":    reason,
		"timestamp": time.Now(),
	})

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

func (s *service) ListDriverProfiles(ctx context.Context, filters map[string]interface{}, page, limit int) (map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	drivers, total, err := s.drvRepo.ListDriverProfiles(ctx, filters, page, limit)
	if err != nil {
		logger.Error("failed to list driver profiles", "error", err)
		return nil, response.InternalServerError("Failed to fetch driver profiles", err)
	}

	return map[string]interface{}{
		"drivers": drivers,
		"total":   total,
		"page":    page,
		"limit":   limit,
	}, nil
}

func (s *service) ListServiceProviderProfiles(ctx context.Context, filters map[string]interface{}, page, limit int) (map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	providers, total, err := s.spRepo.List(ctx, filters, page, limit)
	if err != nil {
		logger.Error("failed to list service provider profiles", "error", err)
		return nil, response.InternalServerError("Failed to fetch service provider profiles", err)
	}

	return map[string]interface{}{
		"providers": providers,
		"total":     total,
		"page":      page,
		"limit":     limit,
	}, nil
}

func (s *service) publishAdminEvent(ctx context.Context, eventType notifications.EventType, data map[string]interface{}) {
	if s.eventProducer == nil {
		return
	}

	go func() {
		if err := s.eventProducer.PublishEvent(ctx, eventType, data); err != nil {
			logger.Error("failed to publish admin event", "error", err, "eventType", eventType)
		}
	}()
}
