package profile

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/profile/dto"
	"github.com/umar5678/go-backend/internal/utils/codegen"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type WalletService interface {
	CreditWallet(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error)
}

type Service interface {
	UpdateEmergencyContact(ctx context.Context, userID string, req dto.UpdateEmergencyContactRequest) error
	GenerateReferralCode(ctx context.Context, userID string) (*dto.ReferralInfoResponse, error)
	ApplyReferralCode(ctx context.Context, userID string, req dto.ApplyReferralRequest) error
	GetReferralInfo(ctx context.Context, userID string) (*dto.ReferralInfoResponse, error)
	SubmitKYC(ctx context.Context, userID string, req dto.SubmitKYCRequest) (*dto.KYCResponse, error)
	GetKYC(ctx context.Context, userID string) (*dto.KYCResponse, error)
	SaveLocation(ctx context.Context, userID string, req dto.SaveLocationRequest) (*dto.SavedLocationResponse, error)
	GetSavedLocations(ctx context.Context, userID string) ([]*dto.SavedLocationResponse, error)
	DeleteLocation(ctx context.Context, userID, locationID string) error
	SetDefaultLocation(ctx context.Context, userID, locationID string) error
	GetRecentLocations(ctx context.Context, userID string) ([]*dto.RecentLocationResponse, error)
}

type service struct {
	repo          Repository
	walletService WalletService
}

func NewService(repo Repository, walletService WalletService) Service {
	return &service{repo: repo, walletService: walletService}
}

func (s *service) UpdateEmergencyContact(ctx context.Context, userID string, req dto.UpdateEmergencyContactRequest) error {
	if err := req.Validate(); err != nil {
		return response.BadRequest(err.Error())
	}

	if err := s.repo.UpdateEmergencyContact(ctx, userID, req.Name, req.Phone); err != nil {
		logger.Error("failed to update emergency contact", "error", err, "userID", userID)
		return response.InternalServerError("Failed to update emergency contact", err)
	}

	logger.Info("emergency contact updated", "userID", userID)
	return nil
}

func (s *service) GenerateReferralCode(ctx context.Context, userID string) (*dto.ReferralInfoResponse, error) {
	code, err := codegen.GenerateReferralCode()
	if err != nil {
		logger.Error("failed to generate referral code", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to generate referral code", err)
	}

	if err := s.repo.GenerateReferralCode(ctx, userID, code); err != nil {
		logger.Error("failed to save referral code", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to save referral code", err)
	}

	count, bonus, err := s.repo.GetReferralStats(ctx, userID)
	if err != nil {
		return nil, response.InternalServerError("Failed to get referral stats", err)
	}

	logger.Info("referral code generated", "userID", userID, "code", code)

	return &dto.ReferralInfoResponse{
		ReferralCode:  code,
		ReferralCount: count,
		ReferralBonus: bonus,
	}, nil
}

func (s *service) ApplyReferralCode(ctx context.Context, userID string, req dto.ApplyReferralRequest) error {
	// Check if code is not empty
	if req.ReferralCode == "" {
		return response.BadRequest("Referral code cannot be empty")
	}

	logger.Info("attempting to apply referral code", "code", req.ReferralCode, "userID", userID)

	// Check if code exists
	referrer, err := s.repo.FindUserByReferralCode(ctx, req.ReferralCode)
	if err != nil {
		logger.Error("referral code lookup failed", "error", err, "code", req.ReferralCode)
		
		// Query all users to debug
		logger.Info("DEBUG: Checking if any user has this code in database")
		return response.BadRequest("Invalid referral code - code does not exist")
	}

	if referrer == nil || referrer.ID == "" {
		logger.Error("referral code lookup returned nil or empty user", "code", req.ReferralCode)
		return response.BadRequest("Invalid referral code")
	}

	if referrer.ID == userID {
		return response.BadRequest("You cannot use your own referral code")
	}

	if err := s.repo.ApplyReferralCode(ctx, userID, req.ReferralCode); err != nil {
		logger.Error("failed to apply referral code", "error", err, "userID", userID)
		return response.InternalServerError("Failed to apply referral code", err)
	}

	// Credit both users with referral bonus
	bonusAmount := 5.0
	metadata := map[string]interface{}{"referral_code": req.ReferralCode}

	// Credit new user
	if _, err := s.walletService.CreditWallet(ctx, userID, bonusAmount, "referral_bonus", referrer.ID, "Referral bonus from code "+req.ReferralCode, metadata); err != nil {
		logger.Error("failed to credit bonus to user", "error", err, "userID", userID)
		// Don't fail the whole request, log it but continue
	}

	// Credit referrer
	if _, err := s.walletService.CreditWallet(ctx, referrer.ID, bonusAmount, "referral_reward", userID, "Referral reward from user "+userID, metadata); err != nil {
		logger.Error("failed to credit bonus to referrer", "error", err, "referrerID", referrer.ID)
		// Don't fail the whole request, log it but continue
	}

	logger.Info("referral code applied successfully", "userID", userID, "referrerID", referrer.ID, "code", req.ReferralCode, "bonusAmount", bonusAmount)
	return nil
}

func (s *service) GetReferralInfo(ctx context.Context, userID string) (*dto.ReferralInfoResponse, error) {
	count, bonus, err := s.repo.GetReferralStats(ctx, userID)
	if err != nil {
		return nil, response.InternalServerError("Failed to get referral info", err)
	}

	// Get user's referral code
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, response.InternalServerError("Failed to get user info", err)
	}

	return &dto.ReferralInfoResponse{
		ReferralCode:  user.ReferralCode,
		ReferralCount: count,
		ReferralBonus: bonus,
	}, nil
}

func (s *service) SubmitKYC(ctx context.Context, userID string, req dto.SubmitKYCRequest) (*dto.KYCResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if KYC already exists
	existing, err := s.repo.FindKYCByUserID(ctx, userID)
	if err == nil && existing.Status == "approved" {
		return nil, response.BadRequest("KYC already approved")
	}

	kyc := &models.UserKYC{
		UserID:        userID,
		IDType:        req.IDType,
		IDNumber:      req.IDNumber,
		IDDocumentURL: req.IDDocumentURL,
		SelfieURL:     req.SelfieURL,
		Status:        "pending",
	}

	if err := s.repo.CreateKYC(ctx, kyc); err != nil {
		logger.Error("failed to create KYC", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to submit KYC", err)
	}

	logger.Info("KYC submitted", "userID", userID, "kycID", kyc.ID)
	return dto.ToKYCResponse(kyc), nil
}

func (s *service) GetKYC(ctx context.Context, userID string) (*dto.KYCResponse, error) {
	kyc, err := s.repo.FindKYCByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("KYC")
	}

	return dto.ToKYCResponse(kyc), nil
}

func (s *service) SaveLocation(ctx context.Context, userID string, req dto.SaveLocationRequest) (*dto.SavedLocationResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	location := &models.SavedLocation{
		UserID:     userID,
		Label:      req.Label,
		CustomName: req.CustomName,
		Address:    req.Address,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		IsDefault:  req.IsDefault,
	}

	if err := s.repo.CreateLocation(ctx, location); err != nil {
		logger.Error("failed to save location", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to save location", err)
	}

	// If this is set as default, update others
	if req.IsDefault {
		s.repo.SetDefaultLocation(ctx, userID, location.ID)
	}

	logger.Info("location saved", "userID", userID, "locationID", location.ID)
	return dto.ToSavedLocationResponse(location), nil
}

func (s *service) GetSavedLocations(ctx context.Context, userID string) ([]*dto.SavedLocationResponse, error) {
	locations, err := s.repo.FindLocationsByUserID(ctx, userID)
	if err != nil {
		return nil, response.InternalServerError("Failed to get saved locations", err)
	}

	result := make([]*dto.SavedLocationResponse, len(locations))
	for i, loc := range locations {
		result[i] = dto.ToSavedLocationResponse(loc)
	}

	return result, nil
}

func (s *service) DeleteLocation(ctx context.Context, userID, locationID string) error {
	location, err := s.repo.FindLocationByID(ctx, locationID)
	if err != nil {
		return response.NotFoundError("Location")
	}

	if location.UserID != userID {
		return response.ForbiddenError("You can only delete your own locations")
	}

	if err := s.repo.DeleteLocation(ctx, locationID); err != nil {
		return response.InternalServerError("Failed to delete location", err)
	}

	logger.Info("location deleted", "userID", userID, "locationID", locationID)
	return nil
}

func (s *service) SetDefaultLocation(ctx context.Context, userID, locationID string) error {
	location, err := s.repo.FindLocationByID(ctx, locationID)
	if err != nil {
		return response.NotFoundError("Location")
	}

	if location.UserID != userID {
		return response.ForbiddenError("Unauthorized")
	}

	if err := s.repo.SetDefaultLocation(ctx, userID, locationID); err != nil {
		return response.InternalServerError("Failed to set default location", err)
	}

	logger.Info("default location set", "userID", userID, "locationID", locationID)
	return nil
}

func (s *service) GetRecentLocations(ctx context.Context, userID string) ([]*dto.RecentLocationResponse, error) {
	locations, err := s.repo.GetRecentLocations(ctx, userID, 10)
	if err != nil {
		return nil, response.InternalServerError("Failed to get recent locations", err)
	}

	result := make([]*dto.RecentLocationResponse, len(locations))
	for i, loc := range locations {
		result[i] = &dto.RecentLocationResponse{
			Address:   loc.Address,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			LastUsed:  loc.CreatedAt,
		}
	}

	return result, nil
}
