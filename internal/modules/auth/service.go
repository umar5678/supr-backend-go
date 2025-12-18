package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/models"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
	"github.com/umar5678/go-backend/internal/modules/riders"
	"github.com/umar5678/go-backend/internal/modules/serviceproviders"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/jwt"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/password"
	"github.com/umar5678/go-backend/internal/utils/response"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	// Phone-based auth (riders/drivers)
	PhoneSignup(ctx context.Context, req authdto.PhoneSignupRequest) (*authdto.AuthResponse, error)
	PhoneLogin(ctx context.Context, req authdto.PhoneLoginRequest) (*authdto.AuthResponse, error)

	// Email-based auth (other roles)
	EmailSignup(ctx context.Context, req authdto.EmailSignupRequest) (*authdto.AuthResponse, error)
	EmailLogin(ctx context.Context, req authdto.EmailLoginRequest) (*authdto.AuthResponse, error)

	// Common
	RefreshToken(ctx context.Context, refreshToken string) (*authdto.AuthResponse, error)
	Logout(ctx context.Context, userID, refreshToken string) error
	GetProfile(ctx context.Context, userID string) (*authdto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID string, req authdto.UpdateProfileRequest) (*authdto.UserResponse, error)
}

type service struct {
	repo                   Repository
	cfg                    *config.Config
	riderService           riders.Service
	serviceProviderService serviceproviders.Service // ✅ ADDED
}

func NewService(
	repo Repository,
	cfg *config.Config,
	riderService riders.Service,
	serviceProviderService serviceproviders.Service, // ✅ ADDED
) Service {
	return &service{
		repo:                   repo,
		cfg:                    cfg,
		riderService:           riderService,
		serviceProviderService: serviceProviderService, // ✅ ADDED
	}
}

// PhoneSignup handles rider/driver signup
func (s *service) PhoneSignup(ctx context.Context, req authdto.PhoneSignupRequest) (*authdto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if phone already exists
	existingUser, err := s.repo.FindByPhone(ctx, req.Phone)
	if err == nil && existingUser != nil {
		// Phone exists - this is a login
		return s.PhoneLogin(ctx, authdto.PhoneLoginRequest{
			Phone: req.Phone,
			// Role:  req.Role,
		})
	}

	// Create new user
	user := &models.User{
		Name:   req.Name,
		Phone:  &req.Phone,
		Role:   req.Role,
		Status: models.StatusActive,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Error("failed to create user", "error", err, "phone", req.Phone)
		return nil, response.InternalServerError("Failed to create account", err)
	}

	// Create wallet for user
	if err := s.createUserWallet(ctx, user); err != nil {
		logger.Error("failed to create wallet", "error", err, "userId", user.ID)
		// Don't fail signup if wallet creation fails
	}

	// CREATE RIDER PROFILE (ADD THIS)
	if user.Role == models.RoleRider {
		if _, err := s.riderService.CreateProfile(ctx, user.ID); err != nil {
			logger.Error("failed to create rider profile", "error", err, "userId", user.ID)
			// Don't fail signup if profile creation fails
		}
	}

	// if user.Role == models.RoleRider {
	// 	if _, err := s.riderService.CreateProfile(ctx, user.ID); err != nil {
	// 		logger.Error("failed to create rider profile", "error", err, "userId", user.ID)
	// 	}

	// } else if user.Role == models.RoleDriver {
	// 	// You'll need to inject driverService similar to riderService
	// 	if _, err := s.driverService.CreateProfile(ctx, user.ID); err != nil {
	// 		logger.Error("failed to create driver profile", "error", err, "userId", user.ID)
	// 	}
	// }

	// Update last login
	s.repo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	logger.Info("phone signup successful", "userId", user.ID, "phone", req.Phone, "role", req.Role)

	return authResp, nil
}

// PhoneLogin handles rider/driver login
func (s *service) PhoneLogin(ctx context.Context, req authdto.PhoneLoginRequest) (*authdto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	user, err := s.repo.FindByPhone(ctx, req.Phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Phone not found - this is a signup
			return nil, response.NotFoundError("Phone number not registered. Please sign up first.")
		}
		return nil, response.InternalServerError("Failed to find user", err)
	}

	// Verify role matches
	// if user.Role != req.Role {
	// 	return nil, response.BadRequest("Invalid role for this phone number")
	// }

	// Check account status
	if user.Status != models.StatusActive {
		return nil, response.ForbiddenError("Account is not active")
	}

	// Update last login
	s.repo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	logger.Info("phone login successful", "userId", user.ID, "phone", req.Phone)

	return authResp, nil
}

// EmailSignup handles email-based signup for other roles
func (s *service) EmailSignup(ctx context.Context, req authdto.EmailSignupRequest) (*authdto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if email already exists
	_, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, response.ConflictError("Email already registered")
	}

	// Hash password
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, response.InternalServerError("Failed to process password", err)
	}

	// ✅ Set appropriate initial status based on role
	initialStatus := models.StatusActive
	if req.Role != models.RoleAdmin {
		// Service providers need approval
		initialStatus = models.StatusPendingApproval
	}

	// Create user
	user := &models.User{
		Name:     req.Name,
		Email:    &req.Email,
		Password: &hashedPassword,
		Role:     req.Role,
		Status:   initialStatus,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Error("failed to create user", "error", err, "email", req.Email)
		return nil, response.InternalServerError("Failed to create account", err)
	}

	// ✅ Create service provider profile if applicable
	if user.IsServiceProvider() {
		serviceCategory := s.mapRoleToCategory(user.Role)

		if _, err := s.serviceProviderService.CreateProfile(ctx, user.ID, serviceCategory); err != nil {
			logger.Error("failed to create service provider profile", "error", err, "userId", user.ID)
			// Don't fail signup if profile creation fails
		}
	}

	// Create wallet for user (if needed for this role)
	if err := s.createUserWallet(ctx, user); err != nil {
		logger.Error("failed to create wallet", "error", err, "userId", user.ID)
	}

	// Update last login
	s.repo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	logger.Info("email signup successful", "userId", user.ID, "email", req.Email, "role", req.Role)

	return authResp, nil
}

// // EmailSignup handles email-based signup for other roles
// func (s *service) EmailSignup(ctx context.Context, req authdto.EmailSignupRequest) (*authdto.AuthResponse, error) {
// 	if err := req.Validate(); err != nil {
// 		return nil, response.BadRequest(err.Error())
// 	}

// 	// Check if email already exists
// 	_, err := s.repo.FindByEmail(ctx, req.Email)
// 	if err == nil {
// 		return nil, response.ConflictError("Email already registered")
// 	}

// 	// Hash password
// 	hashedPassword, err := password.Hash(req.Password)
// 	if err != nil {
// 		return nil, response.InternalServerError("Failed to process password", err)
// 	}

// 	// Create user
// 	user := &models.User{
// 		Name:     req.Name,
// 		Email:    &req.Email,
// 		Password: &hashedPassword,
// 		Role:     req.Role,
// 		Status:   models.StatusActive,
// 	}

// 	if err := s.repo.Create(ctx, user); err != nil {
// 		logger.Error("failed to create user", "error", err, "email", req.Email)
// 		return nil, response.InternalServerError("Failed to create account", err)
// 	}

// 	// Create wallet for user (if needed for this role)
// 	if err := s.createUserWallet(ctx, user); err != nil {
// 		logger.Error("failed to create wallet", "error", err, "userId", user.ID)
// 	}

// 	// Update last login
// 	s.repo.UpdateLastLogin(ctx, user.ID)

// 	// Generate tokens
// 	authResp, err := s.generateAuthResponse(user)
// 	if err != nil {
// 		return nil, err
// 	}

// 	logger.Info("email signup successful", "userId", user.ID, "email", req.Email, "role", req.Role)

// 	return authResp, nil
// }

// EmailLogin handles email-based login
func (s *service) EmailLogin(ctx context.Context, req authdto.EmailLoginRequest) (*authdto.AuthResponse, error) {
	fmt.Println("=========================================================service  called")
	logger.Debug("service fun for emial login, ============")
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.UnauthorizedError("Invalid email or password , email or pas")
		}
		return nil, response.InternalServerError("Failed to find user", err)
	}

	// Verify user has password (email-based account)
	if user.Password == nil {
		return nil, response.BadRequest("This account uses phone authentication")
	}

	// if !password.Verify(*user.Password, req.Password) {
	// 	logger.Debug("password verify: ", *user.Password, req.Password)
	// 	return nil, response.UnauthorizedError("Invalid credentials || password validation failed ")
	// }

	if !password.Verify(req.Password, *user.Password) {
		fmt.Println("RAW stored hash: ", *user.Password)
		logger.Debug("RAW stored hash: ", *user.Password)
		fmt.Println("Hash length: ", len(*user.Password))
		logger.Debug("Hash length: ", len(*user.Password))
		fmt.Println("Hash starts with: ", (*user.Password)[:7]) // should be "$2a$" or "$2b$"

		// Test manually with bcrypt
		err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password))
		fmt.Println("bcrypt compare error: ", err)
		logger.Debug("bcrypt error: ", err)

		return nil, response.UnauthorizedError("Invalid credentials")
	}

	// Check account status
	if user.Status != models.StatusActive {
		return nil, response.ForbiddenError("Account is not active")
	}

	// Update last login
	s.repo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	logger.Info("email login successful", "userId", user.ID, "email", req.Email)

	return authResp, nil
}

// ✅ Helper: Map user role to service category
func (s *service) mapRoleToCategory(role models.UserRole) string {
	switch role {
	case models.RoleDeliveryPerson:
		return "delivery"
	case models.RoleHandyman:
		return "handyman"
	case models.RoleServiceProvider:
		return "general_service"
	default:
		return "general"
	}
}

// RefreshToken generates new tokens
func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*authdto.AuthResponse, error) {
	// Validate refresh token
	claims, err := jwt.ValidateToken(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return nil, response.UnauthorizedError("Invalid refresh token")
	}

	// Check if token is blacklisted
	isBlacklisted, _ := cache.Get(ctx, "blacklist:"+refreshToken)
	if isBlacklisted != "" {
		return nil, response.UnauthorizedError("Token has been revoked")
	}

	// Get user
	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, response.NotFoundError("User")
	}

	// Check account status
	if user.Status != models.StatusActive {
		return nil, response.ForbiddenError("Account is not active")
	}

	// Generate new tokens
	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	// Blacklist old refresh token
	cache.Set(ctx, "blacklist:"+refreshToken, "1", time.Duration(s.cfg.JWT.RefreshExpiry)*20)

	logger.Info("token refreshed", "userId", user.ID)

	return authResp, nil
}

// Logout invalidates tokens
func (s *service) Logout(ctx context.Context, userID, refreshToken string) error {
	// Blacklist refresh token
	if refreshToken != "" {
		cache.Set(ctx, "blacklist:"+refreshToken, "1", time.Duration(s.cfg.JWT.RefreshExpiry)*20)
	}

	// Clear user cache
	cache.Delete(ctx, "user:profile:"+userID)

	logger.Info("user logged out", "userId", userID)

	return nil
}

// GetProfile retrieves user profile
func (s *service) GetProfile(ctx context.Context, userID string) (*authdto.UserResponse, error) {
	// Try cache first
	cacheKey := "user:profile:" + userID
	var cachedUser models.User
	err := cache.GetJSON(ctx, cacheKey, &cachedUser)
	if err == nil {
		return authdto.ToUserResponse(&cachedUser), nil
	}

	// Get from database
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("User")
	}

	// Cache for 5 minutes
	cache.SetJSON(ctx, cacheKey, user, 5*time.Minute)

	return authdto.ToUserResponse(user), nil
}

// UpdateProfile updates user profile
func (s *service) UpdateProfile(ctx context.Context, userID string, req authdto.UpdateProfileRequest) (*authdto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("User")
	}

	// Update fields
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		// Check if email is already taken
		existingUser, err := s.repo.FindByEmail(ctx, *req.Email)
		if err == nil && existingUser.ID != userID {
			return nil, response.ConflictError("Email already in use")
		}
		user.Email = req.Email
	}
	if req.ProfilePhotoURL != nil {
		user.ProfilePhotoURL = req.ProfilePhotoURL
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, response.InternalServerError("Failed to update profile", err)
	}

	// Invalidate cache
	cache.Delete(ctx, "user:profile:"+userID)

	logger.Info("profile updated", "userId", userID)

	return authdto.ToUserResponse(user), nil
}

// Helper: Generate auth response with tokens
func (s *service) generateAuthResponse(user *models.User) (*authdto.AuthResponse, error) {
	// Generate access token
	accessToken, err := jwt.GenerateToken(
		user.ID,
		string(user.Role),
		s.cfg.JWT.Secret,
		time.Duration(s.cfg.JWT.AccessExpiry)*7,
	)
	if err != nil {
		return nil, response.InternalServerError("Failed to generate access token", err)
	}

	// Generate refresh token
	refreshToken, err := jwt.GenerateToken(
		user.ID,
		string(user.Role),
		s.cfg.JWT.Secret,
		time.Duration(s.cfg.JWT.RefreshExpiry)*20,
	)
	if err != nil {
		return nil, response.InternalServerError("Failed to generate refresh token", err)
	}

	return &authdto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         authdto.ToUserResponse(user),
	}, nil
}

// ✅ Helper: Update wallet creation logic
func (s *service) createUserWallet(ctx context.Context, user *models.User) error {
	var walletType models.WalletType
	var initialBalance float64

	switch user.Role {
	case models.RoleRider:
		walletType = models.WalletTypeRider
		initialBalance = 1000.00 // Give riders $1000 fake money
	case models.RoleDriver:
		walletType = models.WalletTypeDriver
		initialBalance = 0.00
	case models.RoleServiceProvider, models.RoleHandyman, models.RoleDeliveryPerson:
		walletType = models.WalletTypeServiceProvider
		initialBalance = 0.00
	default:
		// Admins don't need wallets
		return nil
	}

	wallet := &models.Wallet{
		UserID:      user.ID,
		WalletType:  walletType,
		Balance:     initialBalance,
		HeldBalance: 0.00,
		Currency:    "USD",
		IsActive:    true,
	}

	return s.repo.CreateWallet(ctx, wallet)
}
