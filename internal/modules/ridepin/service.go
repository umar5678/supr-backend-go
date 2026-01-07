// internal/modules/ridepin/service.go
package ridepin

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	VerifyRidePIN(ctx context.Context, userID, pin string) error
	GetUserRidePIN(ctx context.Context, userID string) (string, error) // For showing to rider
	RegenerateRidePIN(ctx context.Context, userID string) (string, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) VerifyRidePIN(ctx context.Context, userID, pin string) error {
	if pin == "" {
		return response.BadRequest("Ride PIN is required")
	}

	if len(pin) != 4 {
		return response.BadRequest("Ride PIN must be 4 digits")
	}

	isValid, err := s.repo.VerifyRidePIN(ctx, userID, pin)
	if err != nil {
		logger.Error("failed to verify ride PIN", "error", err, "userID", userID)
		return response.InternalServerError("Failed to verify PIN", err)
	}

	if !isValid {
		logger.Warn("invalid ride PIN attempt", "userID", userID)
		return response.BadRequest("Invalid Ride PIN")
	}

	logger.Info("ride PIN verified", "userID", userID)
	return nil
}

func (s *service) GetUserRidePIN(ctx context.Context, userID string) (string, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", response.NotFoundError("User")
	}

	return user.RidePIN, nil
}

func (s *service) RegenerateRidePIN(ctx context.Context, userID string) (string, error) {
	newPIN := generateRidePIN()

	if err := s.repo.UpdateRidePIN(ctx, userID, newPIN); err != nil {
		logger.Error("failed to regenerate ride PIN", "error", err, "userID", userID)
		return "", response.InternalServerError("Failed to regenerate PIN", err)
	}

	logger.Info("ride PIN regenerated", "userID", userID)
	return newPIN, nil
}

// Generate random 4-digit PIN
func generateRidePIN() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("%04d", n.Int64())
}
