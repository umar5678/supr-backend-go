package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/umar5678/go-backend/internal/modules/notifications/dto"
	"github.com/umar5678/go-backend/internal/modules/notifications/repository"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type NotificationService interface {
	GetUserNotifications(ctx context.Context, userID uuid.UUID, req *dto.GetNotificationsRequest) (*dto.GetNotificationsResponse, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	DeleteNotification(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, req *dto.GetNotificationsRequest) (*dto.GetNotificationsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	logger.Info("GetUserNotifications querying database", "userID", userID.String(), "page", req.Page, "pageSize", req.PageSize, "offset", offset)

	notifications, total, err := s.repo.GetByUserID(ctx, userID, req.PageSize, offset)
	if err != nil {
		logger.Error("GetUserNotifications database query failed", "error", err, "userID", userID.String())
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	logger.Info("GetUserNotifications database query success", "userID", userID.String(), "count", len(notifications), "total", total)

	notificationDTOs := make([]*dto.NotificationDTO, len(notifications))
	for i, n := range notifications {
		notificationDTOs[i] = dto.ToNotificationDTO(n)
	}

	return &dto.GetNotificationsResponse{
		Notifications: notificationDTOs,
		Total:         total,
		Page:          req.Page,
		PageSize:      req.PageSize,
	}, nil
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	unread, err := s.repo.GetUnreadByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread notifications: %w", err)
	}
	return len(unread), nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
	notification, err := s.repo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	if notification.UserID != userID {
		return fmt.Errorf("unauthorized: notification does not belong to user")
	}

	return s.repo.MarkAsRead(ctx, notificationID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) DeleteNotification(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
	notification, err := s.repo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	if notification.UserID != userID {
		return fmt.Errorf("unauthorized: notification does not belong to user")
	}

	return s.repo.Delete(ctx, notificationID)
}
