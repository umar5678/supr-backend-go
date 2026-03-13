package documents

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// ProfileVerificationHelper provides helper functions for profile verification
type ProfileVerificationHelper struct {
	repo Repository
}

// NewProfileVerificationHelper creates a new profile verification helper
func NewProfileVerificationHelper(repo Repository) *ProfileVerificationHelper {
	return &ProfileVerificationHelper{
		repo: repo,
	}
}

// EnsureDriverProfileNotVerified ensures a driver profile is marked as not verified
// This should be called when a driver profile is first created or when documents are rejected
func (h *ProfileVerificationHelper) EnsureDriverProfileNotVerified(ctx context.Context, driverID string) error {
	if err := h.repo.UpdateDriverProfileVerification(ctx, driverID, false); err != nil {
		logger.Error("failed to ensure driver profile not verified", "error", err, "driverID", driverID)
		return err
	}
	logger.Info("driver profile verified status initialized", "driverID", driverID, "isVerified", false)
	return nil
}

// EnsureServiceProviderProfileNotVerified ensures a service provider profile is marked as not verified
// This should be called when a service provider profile is first created or when documents are rejected
func (h *ProfileVerificationHelper) EnsureServiceProviderProfileNotVerified(ctx context.Context, providerID string) error {
	if err := h.repo.UpdateServiceProviderProfileVerification(ctx, providerID, false); err != nil {
		logger.Error("failed to ensure service provider profile not verified", "error", err, "providerID", providerID)
		return err
	}
	logger.Info("service provider profile verified status initialized", "providerID", providerID, "isVerified", false)
	return nil
}

// CheckAndVerifyDriverProfile checks if driver has all required documents verified and updates profile
func (h *ProfileVerificationHelper) CheckAndVerifyDriverProfile(ctx context.Context, driverID string) error {
	verifiedCount, err := h.repo.CountVerifiedDocumentsByDriverID(ctx, driverID)
	if err != nil {
		logger.Error("failed to count verified documents for driver", "error", err, "driverID", driverID)
		return err
	}

	// Required documents for driver: license, registration, insurance, profile-photo (4 documents)
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

// CheckAndVerifyServiceProviderProfile checks if service provider has all required documents verified and updates profile
func (h *ProfileVerificationHelper) CheckAndVerifyServiceProviderProfile(ctx context.Context, providerID string) error {
	verifiedCount, err := h.repo.CountVerifiedDocumentsByServiceProviderID(ctx, providerID)
	if err != nil {
		logger.Error("failed to count verified documents for service provider", "error", err, "providerID", providerID)
		return err
	}

	// Required documents for service provider: trade-license, profile-photo (2 documents)
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

// InitializeNewDriverProfile initializes a new driver profile with proper defaults
// This should be called immediately after creating a driver profile during registration
func InitializeNewDriverProfile(driverID string) *models.DriverProfile {
	return &models.DriverProfile{
		ID:         driverID,
		IsVerified: false, // Drivers must submit and have documents verified
		Rating:     5.0,   // Default rating
	}
}

// InitializeNewServiceProviderProfile initializes a new service provider profile with proper defaults
// This should be called immediately after creating a service provider profile during registration
func InitializeNewServiceProviderProfile(providerID string) *models.ServiceProviderProfile {
	return &models.ServiceProviderProfile{
		ID:         providerID,
		IsVerified: false, // Service providers must submit and have documents verified
	}
}
