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
	"github.com/umar5678/go-backend/internal/utils/codegen"
	"github.com/umar5678/go-backend/internal/utils/jwt"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/password"
	"github.com/umar5678/go-backend/internal/utils/response"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	PhoneSignup(ctx context.Context, req authdto.PhoneSignupRequest) (*authdto.AuthResponse, error)
	PhoneLogin(ctx context.Context, req authdto.PhoneLoginRequest) (*authdto.AuthResponse, error)

	EmailSignup(ctx context.Context, req authdto.EmailSignupRequest) (*authdto.AuthResponse, error)
	EmailLogin(ctx context.Context, req authdto.EmailLoginRequest) (*authdto.AuthResponse, error)

	RefreshToken(ctx context.Context, refreshToken string) (*authdto.AuthResponse, error)
	Logout(ctx context.Context, userID, refreshToken string) error
	GetProfile(ctx context.Context, userID string) (*authdto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID string, req authdto.UpdateProfileRequest) (*authdto.UserResponse, error)
}

type service struct {
	repo                   Repository
	cfg                    *config.Config
	riderService           riders.Service
	serviceProviderService serviceproviders.Service
}

func NewService(
	repo Repository,
	cfg *config.Config,
	riderService riders.Service,
	serviceProviderService serviceproviders.Service,
) Service {
	return &service{
		repo:                   repo,
		cfg:                    cfg,
		riderService:           riderService,
		serviceProviderService: serviceProviderService,
	}
}

func (s *service) PhoneSignup(ctx context.Context, req authdto.PhoneSignupRequest) (*authdto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	existingUser, err := s.repo.FindByPhone(ctx, req.Phone)
	if err == nil && existingUser != nil {

		if existingUser.Role == req.Role {
			return s.PhoneLogin(ctx, authdto.PhoneLoginRequest{
				Phone: req.Phone,
			})
		}
		existingUser.Role = req.Role
		if err := s.repo.Update(ctx, existingUser); err != nil {
			logger.Error("failed to update user role", "error", err, "phone", req.Phone, "newRole", req.Role)
			return nil, response.InternalServerError("Failed to add new role", err)
		}
		logger.Info("user role updated", "userId", existingUser.ID, "phone", req.Phone, "newRole", req.Role)
		s.repo.UpdateLastLogin(ctx, existingUser.ID)
		authResp, err := s.generateAuthResponse(existingUser)
		if err != nil {
			return nil, err
		}
		return authResp, nil
	}

	referralCode, err := codegen.GenerateReferralCode()
	if err != nil {
		logger.Error("failed to generate referral code", "error", err, "phone", req.Phone)
		return nil, response.InternalServerError("Failed to generate referral code", err)
	}

	ridePIN, err := codegen.GenerateRidePIN()
	if err != nil {
		logger.Error("failed to generate ride PIN", "error", err, "phone", req.Phone)
		return nil, response.InternalServerError("Failed to generate ride PIN", err)
	}

	user := &models.User{
		Name:         req.Name,
		Phone:        &req.Phone,
		Role:         req.Role,
		Status:       models.StatusActive,
		ReferralCode: &referralCode,
		RidePIN:      ridePIN,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Error("failed to create user", "error", err, "phone", req.Phone)
		return nil, response.InternalServerError("Failed to create account", err)
	}

	logger.Info("user created during signup", "userId", user.ID, "phone", req.Phone, "referralCode", user.ReferralCode)

	if err := s.createUserWallet(ctx, user); err != nil {
		logger.Error("failed to create wallet", "error", err, "userId", user.ID)

	}

	if user.Role == models.RoleRider {
		if _, err := s.riderService.CreateProfile(ctx, user.ID); err != nil {
			logger.Error("failed to create rider profile", "error", err, "userId", user.ID)
		}
	}

	s.repo.UpdateLastLogin(ctx, user.ID)

	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	logger.Info("phone signup successful", "userId", user.ID, "phone", req.Phone, "role", req.Role)

	return authResp, nil
}

func (s *service) PhoneLogin(ctx context.Context, req authdto.PhoneLoginRequest) (*authdto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	user, err := s.repo.FindByPhone(ctx, req.Phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {

			return nil, response.NotFoundError("Phone number not registered. Please sign up first.")
		}
		return nil, response.InternalServerError("Failed to find user", err)
	}


	if user.Status != models.StatusActive {
		return nil, response.ForbiddenError("Account is not active")
	}

	s.repo.UpdateLastLogin(ctx, user.ID)


	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	logger.Info("phone login successful", "userId", user.ID, "phone", req.Phone)

	return authResp, nil
}

func (s *service) EmailSignup(ctx context.Context, req authdto.EmailSignupRequest) (*authdto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}


	existingByEmail, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil && existingByEmail != nil {

		if existingByEmail.Role == req.Role {
			return nil, response.ConflictError("Email already registered for this role")
		}

		existingByEmail.Role = req.Role
		if err := s.repo.Update(ctx, existingByEmail); err != nil {
			logger.Error("failed to update user role", "error", err, "email", req.Email, "newRole", req.Role)
			return nil, response.InternalServerError("Failed to add new role", err)
		}
		logger.Info("user role updated", "userId", existingByEmail.ID, "email", req.Email, "newRole", req.Role)

		s.repo.UpdateLastLogin(ctx, existingByEmail.ID)
		authResp, err := s.generateAuthResponse(existingByEmail)
		if err != nil {
			return nil, err
		}
		return authResp, nil
	}

	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, response.InternalServerError("Failed to process password", err)
	}

	initialStatus := models.StatusActive
	if req.Role != models.RoleAdmin {

		initialStatus = models.StatusPendingApproval
	}


	referralCode, err := codegen.GenerateReferralCode()
	if err != nil {
		logger.Error("failed to generate referral code", "error", err, "email", req.Email)
		return nil, response.InternalServerError("Failed to generate referral code", err)
	}


	ridePIN, err := codegen.GenerateRidePIN()
	if err != nil {
		logger.Error("failed to generate ride PIN", "error", err, "email", req.Email)
		return nil, response.InternalServerError("Failed to generate ride PIN", err)
	}

	user := &models.User{
		Name:         req.Name,
		Email:        &req.Email,
		Password:     &hashedPassword,
		Role:         req.Role,
		Status:       initialStatus,
		ReferralCode: &referralCode,
		RidePIN:      ridePIN,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Error("failed to create user", "error", err, "email", req.Email)
		return nil, response.InternalServerError("Failed to create account", err)
	}

	logger.Info("user created during email signup", "userId", user.ID, "email", req.Email, "referralCode", user.ReferralCode)

	if user.IsServiceProvider() {
		serviceCategory := s.mapRoleToCategory(user.Role)

		if _, err := s.serviceProviderService.CreateProfile(ctx, user.ID, serviceCategory); err != nil {
			logger.Error("failed to create service provider profile", "error", err, "userId", user.ID)

		}
	}


	if err := s.createUserWallet(ctx, user); err != nil {
		logger.Error("failed to create wallet", "error", err, "userId", user.ID)
	}


	s.repo.UpdateLastLogin(ctx, user.ID)


	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	logger.Info("email signup successful", "userId", user.ID, "email", req.Email, "role", req.Role)

	return authResp, nil
}

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

	if user.Password == nil {
		return nil, response.BadRequest("This account uses phone authentication")
	}

	if !password.Verify(req.Password, *user.Password) {
		fmt.Println("RAW stored hash: ", *user.Password)
		logger.Debug("RAW stored hash: ", *user.Password)
		fmt.Println("Hash length: ", len(*user.Password))
		logger.Debug("Hash length: ", len(*user.Password))
		fmt.Println("Hash starts with: ", (*user.Password)[:7])

		err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password))
		fmt.Println("bcrypt compare error: ", err)
		logger.Debug("bcrypt error: ", err)

		return nil, response.UnauthorizedError("Invalid credentials")
	}

	if user.Status != models.StatusActive {
		return nil, response.ForbiddenError("Account is not active")
	}

	s.repo.UpdateLastLogin(ctx, user.ID)

	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	logger.Info("email login successful", "userId", user.ID, "email", req.Email)

	return authResp, nil
}

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

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*authdto.AuthResponse, error) {

	claims, err := jwt.ValidateToken(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return nil, response.UnauthorizedError("Invalid refresh token")
	}

	isBlacklisted, _ := cache.Get(ctx, "blacklist:"+refreshToken)
	if isBlacklisted != "" {
		return nil, response.UnauthorizedError("Token has been revoked")
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, response.NotFoundError("User")
	}

	if user.Status != models.StatusActive {
		return nil, response.ForbiddenError("Account is not active")
	}

	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		return nil, err
	}

	cache.Set(ctx, "blacklist:"+refreshToken, "1", time.Duration(s.cfg.JWT.RefreshExpiry)*20)

	logger.Info("token refreshed", "userId", user.ID)

	return authResp, nil
}

func (s *service) Logout(ctx context.Context, userID, refreshToken string) error {

	if refreshToken != "" {
		cache.Set(ctx, "blacklist:"+refreshToken, "1", time.Duration(s.cfg.JWT.RefreshExpiry)*20)
	}


	cache.Delete(ctx, "user:profile:"+userID)

	logger.Info("user logged out", "userId", userID)

	return nil
}


func (s *service) GetProfile(ctx context.Context, userID string) (*authdto.UserResponse, error) {

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("User")
	}

	cacheKey := "user:profile:" + userID
	cache.SetJSON(ctx, cacheKey, user, 5*time.Minute)

	return authdto.ToUserResponse(user), nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, req authdto.UpdateProfileRequest) (*authdto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("User")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {

		existingUser, err := s.repo.FindByEmail(ctx, *req.Email)
		if err == nil && existingUser.ID != userID {
			return nil, response.ConflictError("Email already in use")
		}
		user.Email = req.Email
	}
	if req.ProfilePhotoURL != nil {
		user.ProfilePhotoURL = req.ProfilePhotoURL
	}

	if req.Phone != nil {

		existingUser, err := s.repo.FindByPhone(ctx, *req.Phone)
		if err == nil && existingUser.ID != userID {
			return nil, response.ConflictError("Phone number already in use")
		}
		user.Phone = req.Phone
	}

	if req.Gender != nil {
		user.Gender = req.Gender
	}

	if req.DOB != nil {
		dob, err := req.ParseDOB()
		if err != nil {
			return nil, response.BadRequest("Invalid date of birth")
		}
		user.DOB = dob
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, response.InternalServerError("Failed to update profile", err)
	}

	cache.Delete(ctx, "user:profile:"+userID)

	logger.Info("profile updated", "userId", userID)

	return authdto.ToUserResponse(user), nil
}

func (s *service) generateAuthResponse(user *models.User) (*authdto.AuthResponse, error) {

	accessToken, err := jwt.GenerateToken(
		user.ID,
		string(user.Role),
		s.cfg.JWT.Secret,
		time.Duration(s.cfg.JWT.AccessExpiry)*7,
	)
	if err != nil {
		return nil, response.InternalServerError("Failed to generate access token", err)
	}

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

func (s *service) createUserWallet(ctx context.Context, user *models.User) error {
	var walletType models.WalletType
	var initialBalance float64

	switch user.Role {
	case models.RoleRider:
		walletType = models.WalletTypeRider
		initialBalance = 1000.00
	case models.RoleDriver:
		walletType = models.WalletTypeDriver
		initialBalance = 0.00
	case models.RoleServiceProvider, models.RoleHandyman, models.RoleDeliveryPerson:
		walletType = models.WalletTypeServiceProvider
		initialBalance = 0.00
	default:
		return nil
	}

	wallet := &models.Wallet{
		UserID:      user.ID,
		WalletType:  walletType,
		Balance:     initialBalance,
		HeldBalance: 0.00,
		Currency:    "INR",
		IsActive:    true,
	}

	return s.repo.CreateWallet(ctx, wallet)
}
