package documents

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type ProfileVerificationHelper struct {
	repo Repository
}

func NewProfileVerificationHelper(repo Repository) *ProfileVerificationHelper {
	return &ProfileVerificationHelper{
		repo: repo,
	}
}

func (h *ProfileVerificationHelper) EnsureDriverProfileNotVerified(ctx context.Context, driverID string) error {
	if err := h.repo.UpdateDriverProfileVerification(ctx, driverID, false); err != nil {
		logger.Error("failed to ensure driver profile not verified", "error", err, "driverID", driverID)
		return err
	}
	logger.Info("driver profile verified status initialized", "driverID", driverID, "isVerified", false)
	return nil
}

func (h *ProfileVerificationHelper) EnsureServiceProviderProfileNotVerified(ctx context.Context, providerID string) error {
	if err := h.repo.UpdateServiceProviderProfileVerification(ctx, providerID, false); err != nil {
		logger.Error("failed to ensure service provider profile not verified", "error", err, "providerID", providerID)
		return err
	}
	logger.Info("service provider profile verified status initialized", "providerID", providerID, "isVerified", false)
	return nil
}

func (h *ProfileVerificationHelper) CheckAndVerifyDriverProfile(ctx context.Context, driverID string) error {
	verifiedCount, err := h.repo.CountVerifiedDocumentsByDriverID(ctx, driverID)
	if err != nil {
		logger.Error("failed to count verified documents for driver", "error", err, "driverID", driverID)
		return err
	}

	requiredDocsCount := 4
	isVerified := verifiedCount >= requiredDocsCount

	if err := h.repo.UpdateDriverProfileVerification(ctx, driverID, isVerified); err != nil {
		logger.Error("failed to update driver profile verification", "error", err, "driverID", driverID)
		return err
	}

	logger.Info("driver profile verification check completed",
		"driverID", driverID,
		"verifiedDocuments", verifiedCount,
		"requiredDocuments", requiredDocsCount,
		"isVerified", isVerified)

	return nil
}

func (h *ProfileVerificationHelper) CheckAndVerifyServiceProviderProfile(ctx context.Context, providerID string) error {
	verifiedCount, err := h.repo.CountVerifiedDocumentsByServiceProviderID(ctx, providerID)
	if err != nil {
		logger.Error("failed to count verified documents for service provider", "error", err, "providerID", providerID)
		return err
	}

	requiredDocsCount := 2
	isVerified := verifiedCount >= requiredDocsCount

	if err := h.repo.UpdateServiceProviderProfileVerification(ctx, providerID, isVerified); err != nil {
		logger.Error("failed to update service provider profile verification", "error", err, "providerID", providerID)
		return err
	}

	logger.Info("service provider profile verification check completed",
		"providerID", providerID,
		"verifiedDocuments", verifiedCount,
		"requiredDocuments", requiredDocsCount,
		"isVerified", isVerified)

	return nil
}

func InitializeNewDriverProfile(driverID string) *models.DriverProfile {
	return &models.DriverProfile{
		ID:         driverID,
		IsVerified: false, 
		Rating:     5.0,   
	}
}

func InitializeNewServiceProviderProfile(providerID string) *models.ServiceProviderProfile {
	return &models.ServiceProviderProfile{
		ID:         providerID,
		IsVerified: false,
	}
}
