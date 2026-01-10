// internal/modules/drivers/service.go
package drivers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	driverdto "github.com/umar5678/go-backend/internal/modules/drivers/dto"
	walletservice "github.com/umar5678/go-backend/internal/modules/wallet"
	walletdto "github.com/umar5678/go-backend/internal/modules/wallet/dto"

	// trackingDto "github.com/umar5678/go-backend/internal/modules/tracking/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/helpers"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
	"github.com/umar5678/go-backend/internal/websocket/websocketutils"
	"gorm.io/gorm"
)

type Service interface {
	RegisterDriver(ctx context.Context, userID string, req driverdto.RegisterDriverRequest) (*driverdto.DriverProfileResponse, error)
	GetProfile(ctx context.Context, userID string) (*driverdto.DriverProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req driverdto.UpdateDriverProfileRequest) (*driverdto.DriverProfileResponse, error)
	UpdateVehicle(ctx context.Context, userID string, req driverdto.UpdateVehicleRequest) (*driverdto.VehicleResponse, error)
	UpdateStatus(ctx context.Context, userID string, req driverdto.UpdateStatusRequest) (*driverdto.DriverProfileResponse, error)
	// UpdateLocation(ctx context.Context, userID string, req driverdto.UpdateLocationRequest) error
	GetWallet(ctx context.Context, userID string) (*driverdto.WalletResponse, error)
	GetDashboard(ctx context.Context, userID string) (*driverdto.DriverDashboardResponse, error)
	
	// ✅ Wallet management methods
	TopUpWallet(ctx context.Context, userID string, req driverdto.WalletTopUpRequest) (*driverdto.WalletTopUpResponse, error)
	GetWalletStatus(ctx context.Context, userID string) (*driverdto.WalletStatusResponse, error)
	GetWalletTransactionHistory(ctx context.Context, userID string) ([]*driverdto.WalletTransactionResponse, error)

	// CreateDriverProfile(ctx context.Context, userID string, req driverdto.CreateDriverProfileRequest) (*driverdto.DriverProfileResponse, error)
	// GetDriverProfile(ctx context.Context, userID string) (*driverdto.DriverProfileResponse, error)
	// UpdateDriverProfile(ctx context.Context, userID string, req driverdto.UpdateDriverProfileRequest) (*driverdto.DriverProfileResponse, error)
	// UpdateDriverStatus(ctx context.Context, userID string, req driverdto.UpdateDriverStatusRequest) (*driverdto.DriverProfileResponse, error)
	UpdateLocation(ctx context.Context, userID string, req driverdto.UpdateLocationRequest) error
	// GetDriverStats(ctx context.Context, userID string) (*trackingDto.DriverStatsResponse, error)
}

type service struct {
	repo          Repository
	walletService walletservice.Service
	db            *gorm.DB
}

func NewService(repo Repository, walletService walletservice.Service, db *gorm.DB) Service {
	return &service{
		repo:          repo,
		walletService: walletService,
		db:            db,
	}
}

func (s *service) RegisterDriver(ctx context.Context, userID string, req driverdto.RegisterDriverRequest) (*driverdto.DriverProfileResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if user is already a driver
	_, err := s.repo.FindDriverByUserID(ctx, userID)
	if err == nil {
		return nil, response.BadRequest("User is already registered as a driver")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("failed to check existing driver", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to check driver status", err)
	}

	// Check if license number already exists
	_, err = s.repo.FindDriverByLicense(ctx, req.LicenseNumber)
	if err == nil {
		return nil, response.BadRequest("License number is already registered")
	}

	// Check if license plate already exists
	_, err = s.repo.FindVehicleByPlate(ctx, req.Vehicle.LicensePlate)
	if err == nil {
		return nil, response.BadRequest("License plate is already registered")
	}

	// Create driver profile
	driver := &models.DriverProfile{
		UserID:        userID,
		LicenseNumber: req.LicenseNumber,
		Status:        "online",
		Rating:        5.0,
		IsVerified:    true, // Auto-approve (no document verification needed)
	}

	if err := s.repo.CreateDriver(ctx, driver); err != nil {
		logger.Error("failed to create driver profile", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create driver profile", err)
	}

	// Create vehicle
	vehicle := &models.Vehicle{
		DriverID:      driver.ID,
		VehicleTypeID: req.Vehicle.VehicleTypeID,
		Make:          req.Vehicle.Make,
		Model:         req.Vehicle.Model,
		Year:          req.Vehicle.Year,
		Color:         req.Vehicle.Color,
		LicensePlate:  req.Vehicle.LicensePlate,
		Capacity:      req.Vehicle.Capacity,
		IsActive:      true,
	}

	if err := s.repo.CreateVehicle(ctx, vehicle); err != nil {
		logger.Error("failed to create vehicle", "error", err, "driverID", driver.ID)
		return nil, response.InternalServerError("Failed to register vehicle", err)
	}

	// Fetch complete profile with relations
	driver, _ = s.repo.FindDriverByID(ctx, driver.ID)

	logger.Info("driver registered successfully",
		"driverID", driver.ID,
		"userID", userID,
		"vehicleType", req.Vehicle.VehicleTypeID,
	)

	return driverdto.ToDriverProfileResponse(driver), nil
}

func (s *service) GetProfile(ctx context.Context, userID string) (*driverdto.DriverProfileResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("driver:profile:%s", userID)
	var cached models.DriverProfile

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		logger.Debug("driver profile cache hit", "userID", userID)
		return driverdto.ToDriverProfileResponse(&cached), nil
	}

	// Get from database
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("Driver profile")
		}
		logger.Error("failed to fetch driver profile", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to fetch profile", err)
	}

	// Cache for 5 minutes
	cache.SetJSON(ctx, cacheKey, driver, 5*time.Minute)

	return driverdto.ToDriverProfileResponse(driver), nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, req driverdto.UpdateDriverProfileRequest) (*driverdto.DriverProfileResponse, error) {
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	// Update fields
	if req.LicenseNumber != nil {
		// Check if new license number is unique
		existing, err := s.repo.FindDriverByLicense(ctx, *req.LicenseNumber)
		if err == nil && existing.ID != driver.ID {
			return nil, response.BadRequest("License number is already in use")
		}
		driver.LicenseNumber = *req.LicenseNumber
	}

	if err := s.repo.UpdateDriver(ctx, driver); err != nil {
		logger.Error("failed to update driver profile", "error", err, "driverID", driver.ID)
		return nil, response.InternalServerError("Failed to update profile", err)
	}

	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("driver:profile:%s", userID))

	// Fetch updated profile
	driver, _ = s.repo.FindDriverByID(ctx, driver.ID)

	logger.Info("driver profile updated", "driverID", driver.ID, "userID", userID)

	return driverdto.ToDriverProfileResponse(driver), nil
}

func (s *service) UpdateVehicle(ctx context.Context, userID string, req driverdto.UpdateVehicleRequest) (*driverdto.VehicleResponse, error) {
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	vehicle, err := s.repo.FindVehicleByDriverID(ctx, driver.ID)
	if err != nil {
		return nil, response.NotFoundError("Vehicle")
	}

	// Update fields
	if req.VehicleTypeID != nil {
		vehicle.VehicleTypeID = *req.VehicleTypeID
	}
	if req.Make != nil {
		vehicle.Make = *req.Make
	}
	if req.Model != nil {
		vehicle.Model = *req.Model
	}
	if req.Year != nil {
		vehicle.Year = *req.Year
	}
	if req.Color != nil {
		vehicle.Color = *req.Color
	}
	if req.LicensePlate != nil {
		// Check if plate is unique
		existing, err := s.repo.FindVehicleByPlate(ctx, *req.LicensePlate)
		if err == nil && existing.ID != vehicle.ID {
			return nil, response.BadRequest("License plate is already in use")
		}
		vehicle.LicensePlate = *req.LicensePlate
	}
	if req.Capacity != nil {
		vehicle.Capacity = *req.Capacity
	}
	if req.IsActive != nil {
		vehicle.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateVehicle(ctx, vehicle); err != nil {
		logger.Error("failed to update vehicle", "error", err, "vehicleID", vehicle.ID)
		return nil, response.InternalServerError("Failed to update vehicle", err)
	}

	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("driver:profile:%s", userID))

	// Fetch updated vehicle
	vehicle, _ = s.repo.FindVehicleByDriverID(ctx, driver.ID)

	logger.Info("vehicle updated", "vehicleID", vehicle.ID, "driverID", driver.ID)

	return driverdto.ToVehicleResponse(vehicle), nil
}

func (s *service) UpdateStatus(ctx context.Context, userID string, req driverdto.UpdateStatusRequest) (*driverdto.DriverProfileResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	// Check if driver can go online
	if req.Status == "online" {
		if !driver.IsVerified {
			return nil, response.BadRequest("Driver is not verified yet")
		}

		// Check if vehicle exists and is active
		vehicle, err := s.repo.FindVehicleByDriverID(ctx, driver.ID)
		if err != nil {
			return nil, response.BadRequest("Vehicle information is required to go online")
		}
		if !vehicle.IsActive {
			return nil, response.BadRequest("Vehicle is not active")
		}
	}

	oldStatus := driver.Status

	if err := s.repo.UpdateDriverStatus(ctx, driver.ID, req.Status); err != nil {
		logger.Error("failed to update driver status", "error", err, "driverID", driver.ID)
		return nil, response.InternalServerError("Failed to update status", err)
	}

	// Update online status in Redis
	onlineKey := fmt.Sprintf("driver:online:%s", driver.ID)
	if req.Status == "online" {
		cache.Set(ctx, onlineKey, "true", 5*time.Minute) // TTL for heartbeat

		// Add to online drivers set
		cache.SessionClient.SAdd(ctx, "drivers:online", driver.ID)
	} else {
		cache.Delete(ctx, onlineKey)
		cache.SessionClient.SRem(ctx, "drivers:online", driver.ID)
	}

	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("driver:profile:%s", userID))

	// Broadcast status change via WebSocket
	go func() {
		helpers.BroadcastNotification(map[string]interface{}{
			"type":      "driver_status_changed",
			"driverId":  driver.ID,
			"status":    req.Status,
			"oldStatus": oldStatus,
		})
	}()

	// Fetch updated profile
	driver, _ = s.repo.FindDriverByID(ctx, driver.ID)

	logger.Info("driver status updated",
		"driverID", driver.ID,
		"oldStatus", oldStatus,
		"newStatus", req.Status,
	)

	return driverdto.ToDriverProfileResponse(driver), nil
}

// internal/modules/drivers/service.go

// func (s *service) UpdateLocation(ctx context.Context, userID string, req driverdto.UpdateLocationRequest) error {
// 	if err := req.Validate(); err != nil {
// 		return response.BadRequest(err.Error())
// 	}

// 	driver, err := s.repo.FindDriverByUserID(ctx, userID)
// 	if err != nil {
// 		return response.NotFoundError("Driver profile")
// 	}

// 	// ✅ FIXED: Allow location updates for online, busy, AND on_trip drivers
// 	if driver.Status != "online" && driver.Status != "busy" && driver.Status != "on_trip" {
// 		return response.BadRequest("Driver must be online to update location")
// 	}

// 	// Update location in database
// 	if err := s.repo.UpdateDriverLocation(ctx, driver.ID, req.Latitude, req.Longitude, req.Heading); err != nil {
// 		logger.Error("failed to update driver location", "error", err, "driverID", driver.ID)
// 		return response.InternalServerError("Failed to update location", err)
// 	}

// 	// Update location in Redis (for faster access)
// 	locationKey := fmt.Sprintf("driver:location:%s", driver.ID)
// 	locationData := map[string]interface{}{
// 		"latitude":  req.Latitude,
// 		"longitude": req.Longitude,
// 		"heading":   req.Heading,
// 		"timestamp": time.Now().Unix(),
// 	}
// 	cache.SetJSON(ctx, locationKey, locationData, 30*time.Second)

// 	// Refresh online status TTL
// 	onlineKey := fmt.Sprintf("driver:online:%s", driver.ID)
// 	cache.Set(ctx, onlineKey, "true", 5*time.Minute)

// 	logger.Debug("driver location updated",
// 		"driverID", driver.ID,
// 		"userID", userID,
// 		"status", driver.Status,
// 		"lat", req.Latitude,
// 		"lng", req.Longitude,
// 	)

// 	return nil
// }

func (s *service) GetWallet(ctx context.Context, userID string) (*driverdto.WalletResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("driver:wallet:%s", userID)
	var cached driverdto.WalletResponse

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		return &cached, nil
	}

	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	walletResp := &driverdto.WalletResponse{
		Balance:       driver.WalletBalance,
		TotalEarnings: driver.TotalEarnings,
		PendingAmount: 0, // TODO: Calculate from pending rides
	}

	// Cache for 1 minute
	cache.SetJSON(ctx, cacheKey, walletResp, 1*time.Minute)

	return walletResp, nil
}

func (s *service) GetDashboard(ctx context.Context, userID string) (*driverdto.DriverDashboardResponse, error) {
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	// TODO: Calculate today's trips and earnings from rides table
	// For now, returning mock data
	dashboard := &driverdto.DriverDashboardResponse{
		TodayTrips:     0, // TODO: Query from rides
		TodayEarnings:  0, // TODO: Query from rides
		WeekEarnings:   0, // TODO: Query from rides
		Rating:         driver.Rating,
		AcceptanceRate: driver.AcceptanceRate,
		CompletionRate: 100 - driver.CancellationRate,
		TotalTrips:     driver.TotalTrips,
		WalletBalance:  driver.WalletBalance,
		Status:         driver.Status,
		Profile:        driverdto.ToDriverProfileResponse(driver),
	}

	return dashboard, nil
}

// ✅ UPDATED: UpdateLocation with location streaming to rider
func (s *service) UpdateLocation(ctx context.Context, userID string, req driverdto.UpdateLocationRequest) error {
	logger.Info("=====================")
	logger.Info("=============== 📍 DRIVERS MODULE: UpdateLocation CALLED",
		"userID", userID,
		"lat", req.Latitude,
		"lng", req.Longitude,
	)

	if err := req.Validate(); err != nil {
		logger.Error(" ================= ❌ Validation failed", "error", err)
		return response.BadRequest(err.Error())
	}

	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("====================== ❌ Driver profile not found", "userID", userID, "error", err)
		return response.NotFoundError("Driver profile")
	}

	logger.Info("✅ Driver profile found",
		"userID", userID,
		"driverProfileID", driver.ID,
		"driverStatus", driver.Status,
	)

	// Allow location updates for active drivers
	if driver.Status != "online" && driver.Status != "busy" && driver.Status != "on_trip" {
		logger.Warn("===================== ❌ Driver not in active status",
			"driverID", driver.ID,
			"status", driver.Status,
		)
		return response.BadRequest("Driver must be online to update location")
	}

	// Update location in database
	logger.Info("💾 Updating location in database...")
	if err := s.repo.UpdateDriverLocation(ctx, driver.ID, req.Latitude, req.Longitude, req.Heading); err != nil {
		logger.Error("================= ❌ Failed to update driver location", "error", err, "driverID", driver.ID)
		return response.InternalServerError("Failed to update location", err)
	}

	logger.Info("✅ Location saved to database")

	// Update location in Redis (for faster access)
	locationKey := fmt.Sprintf("driver:location:%s", driver.ID)
	locationData := map[string]interface{}{
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"heading":   req.Heading,
		"timestamp": time.Now().Unix(),
	}
	cache.SetJSON(ctx, locationKey, locationData, 30*time.Second)

	logger.Info("✅ Location cached in Redis")

	// Refresh online status TTL
	onlineKey := fmt.Sprintf("driver:online:%s", driver.ID)
	cache.Set(ctx, onlineKey, "true", 5*time.Minute)

	// ✅ NEW: Check for active ride and stream location to rider
	logger.Info("🔍 Checking for active ride...")
	activeRideCacheKey := fmt.Sprintf("driver:active:ride:%s", driver.ID)
	var rideData map[string]string

	err = cache.GetJSON(ctx, activeRideCacheKey, &rideData)
	if err != nil {
		logger.Debug("⚠️ No active ride in cache",
			"driverProfileID", driver.ID,
			"cacheKey", activeRideCacheKey,
			"error", err,
		)
	} else if rideData != nil {
		rideID := rideData["rideID"]
		riderID := rideData["riderID"]

		if rideID != "" && riderID != "" {
			logger.Info("✅ Active ride FOUND",
				"driverProfileID", driver.ID,
				"rideID", rideID,
				"riderUserID", riderID,
			)

			logger.Info("📡 Starting location stream to rider...")

			// Stream location to rider (non-blocking)
			go func() {
				logger.Info("🎯 Goroutine STARTED for location streaming")

				locationUpdateData := map[string]interface{}{
					"rideId":   rideID,
					"driverId": driver.ID,
					"location": map[string]interface{}{
						"latitude":  req.Latitude,
						"longitude": req.Longitude,
						"heading":   req.Heading,
						"speed":     0.0,
						"accuracy":  0.0,
						"timestamp": time.Now(),
					},
					"timestamp": time.Now().UTC(),
				}

				logger.Info("📤 Calling SendRideLocationUpdate",
					"riderUserID", riderID,
					"rideID", rideID,
					"lat", req.Latitude,
					"lng", req.Longitude,
				)

				if err := websocketutils.SendRideLocationUpdate(riderID, locationUpdateData); err != nil {
					logger.Error("============= ❌ Failed to send location update to rider",
						"error", err,
						"riderUserID", riderID,
						"rideID", rideID,
						"driverID", driver.ID,
					)
				} else {
					logger.Info("============= ✅ Location successfully streamed to rider",
						"riderUserID", riderID,
						"rideID", rideID,
					)
				}
			}()
		} else {
			logger.Warn("================⚠️ Active ride data incomplete",
				"rideID", rideID,
				"riderID", riderID,
			)
		}
	}

	logger.Info("✅ UpdateLocation completed")
	logger.Info("============================")

	return nil
}

// ✅ TopUpWallet adds funds to driver wallet and handles account unrestriction
// Production Ready: Only change ProcessPayment() implementation, everything else stays the same
func (s *service) TopUpWallet(ctx context.Context, userID string, req driverdto.WalletTopUpRequest) (*driverdto.WalletTopUpResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Get driver profile to check restriction status
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	logger.Info("processing wallet top-up",
		"userID", userID,
		"driverID", driver.ID,
		"amount", req.Amount,
		"paymentMethod", req.PaymentMethod)

	// ✅ PRODUCTION: Only this function needs to be changed
	// Everything else: Standard wallet operations (no changes needed)
	paymentResult, err := s.ProcessPayment(ctx, &PaymentRequest{
		DriverID:      driver.ID,
		UserID:        userID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Reference:     req.Reference,
	})
	if err != nil {
		logger.Error("payment processing failed", "error", err, "userID", userID)
		return nil, response.BadRequest(fmt.Sprintf("Payment failed: %v", err))
	}

	if !paymentResult.Success {
		logger.Warn("payment rejected",
			"userID", userID,
			"reason", paymentResult.Error)
		return nil, response.BadRequest(paymentResult.Error)
	}

	// Credit driver wallet through wallet service
	txn, err := s.walletService.CreditDriverWallet(
		ctx,
		userID,
		req.Amount,
		"wallet_topup",
		paymentResult.TransactionID, // Payment provider transaction ID
		fmt.Sprintf("Wallet top-up of $%.2f via %s (Order: %s)", req.Amount, req.PaymentMethod, paymentResult.OrderID),
		map[string]interface{}{
			"payment_method":   req.PaymentMethod,
			"reference":        req.Reference,
			"order_id":         paymentResult.OrderID,
			"transaction_id":   paymentResult.TransactionID,
			"payment_provider": paymentResult.Provider,
		},
	)
	if err != nil {
		logger.Error("failed to credit wallet", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to credit wallet", err)
	}

	// Get updated wallet info through wallet service
	walletInfo, err := s.walletService.GetWallet(ctx, userID)
	if err != nil {
		logger.Error("failed to get wallet after top-up", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to fetch wallet", err)
	}

	// Check if account was restricted and can now be unrestricted
	accountRestricted := driver.IsRestricted
	if driver.IsRestricted && walletInfo.Balance >= 0 {
		// Unrestrict account
		if err := s.walletService.UnrestrictDriverAccount(ctx, driver.ID); err != nil {
			logger.Error("failed to unrestrict account", "error", err, "driverID", driver.ID)
		} else {
			accountRestricted = false
			logger.Info("driver account unrestricted after top-up",
				"driverID", driver.ID,
				"newBalance", walletInfo.Balance)
		}
	}

	response := &driverdto.WalletTopUpResponse{
		TransactionID:    txn.ID,
		Amount:           req.Amount,
		PaymentMethod:    req.PaymentMethod,
		PreviousBalance:  walletInfo.Balance - req.Amount,
		NewBalance:       walletInfo.Balance,
		AccountRestricted: accountRestricted,
		Message:          "Wallet topped up successfully",
		Timestamp:        time.Now(),
	}

	if accountRestricted {
		response.Message = fmt.Sprintf("Wallet topped up. Add $%.2f more to lift account restriction", driver.MinBalanceThreshold*-1)
	}

	logger.Info("wallet top-up completed successfully",
		"userID", userID,
		"driverID", driver.ID,
		"amount", req.Amount,
		"orderId", paymentResult.OrderID,
		"transactionId", paymentResult.TransactionID)

	return response, nil
}

// ============================================================
// PAYMENT PROCESSING - Only file to modify for production
// ============================================================

// PaymentRequest wraps payment details for processing
type PaymentRequest struct {
	DriverID      string
	UserID        string
	Amount        float64
	PaymentMethod string
	Reference     *string
}

// PaymentResult from payment provider
type PaymentResult struct {
	Success       bool   // true if payment successful
	OrderID       string // Order ID from payment provider
	TransactionID string // Transaction ID from payment provider
	Provider      string // Payment provider name (razorpay, stripe, etc.)
	Error         string // Error message if failed
}

// ✅ ProcessPayment - PRODUCTION IMPLEMENTATION
// This is the ONLY function that needs to change for real payment integration
// Current: Mock implementation
// Production: Replace with actual payment provider integration
func (s *service) ProcessPayment(ctx context.Context, payReq *PaymentRequest) (*PaymentResult, error) {
	logger.Info("processing payment",
		"driverID", payReq.DriverID,
		"amount", payReq.Amount,
		"method", payReq.PaymentMethod)

	// ========================================================
	// OPTION 1: RAZORPAY INTEGRATION
	// ========================================================
	// Uncomment to use Razorpay (replace with your implementation)
	/*
	import "github.com/razorpay/razorpay-go"
	
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))
	
	orderData := map[string]interface{}{
		"amount":   int(payReq.Amount * 100), // Amount in paise
		"currency": "INR",
		"receipt":  fmt.Sprintf("topup_%s_%d", payReq.DriverID, time.Now().Unix()),
		"notes": map[string]interface{}{
			"driver_id": payReq.DriverID,
			"user_id":   payReq.UserID,
		},
	}
	
	order, err := client.Order.Create(orderData, nil)
	if err != nil {
		logger.Error("razorpay order creation failed", "error", err)
		return &PaymentResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create order: %v", err),
		}, nil
	}
	
	// TODO: Frontend pays using order.ID
	// Payment confirmation happens via webhook
	return &PaymentResult{
		Success:       true,
		OrderID:       order.ID,
		TransactionID: order.ID, // Update after webhook
		Provider:      "razorpay",
	}, nil
	*/

	// ========================================================
	// OPTION 2: STRIPE INTEGRATION
	// ========================================================
	// Uncomment to use Stripe (replace with your implementation)
	/*
	import "github.com/stripe/stripe-go/v72"
	import "github.com/stripe/stripe-go/v72/paymentintent"
	
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	
	piParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(payReq.Amount * 100)), // Amount in cents
		Currency: stripe.String(string(stripe.CurrencyINR)),
		Metadata: map[string]string{
			"driver_id": payReq.DriverID,
			"user_id":   payReq.UserID,
		},
	}
	
	pi, err := paymentintent.New(piParams)
	if err != nil {
		logger.Error("stripe payment intent creation failed", "error", err)
		return &PaymentResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create payment: %v", err),
		}, nil
	}
	
	return &PaymentResult{
		Success:       true,
		OrderID:       pi.ID,
		TransactionID: pi.ID,
		Provider:      "stripe",
	}, nil
	*/

	// ========================================================
	// OPTION 3: PAYPAL INTEGRATION
	// ========================================================
	// Uncomment to use PayPal (replace with your implementation)
	/*
	import "github.com/plutov/paypal/v4"
	
	client, err := paypal.NewClient(os.Getenv("PAYPAL_CLIENT_ID"), os.Getenv("PAYPAL_CLIENT_SECRET"), paypal.APIBaseSandBox)
	if err != nil {
		logger.Error("paypal client creation failed", "error", err)
		return &PaymentResult{
			Success: false,
			Error:   "Failed to create PayPal client",
		}, nil
	}
	
	orderBody := paypal.OrderRequest{
		Intent: "CAPTURE",
		PurchaseUnits: []paypal.PurchaseUnitRequest{
			{
				ReferenceID: fmt.Sprintf("topup_%s_%d", payReq.DriverID, time.Now().Unix()),
				Amount: &paypal.AmountWithBreakdown{
					Value:    fmt.Sprintf("%.2f", payReq.Amount),
					Currency: "INR",
				},
			},
		},
	}
	
	order, err := client.CreateOrder(ctx, orderBody)
	if err != nil {
		logger.Error("paypal order creation failed", "error", err)
		return &PaymentResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create order: %v", err),
		}, nil
	}
	
	return &PaymentResult{
		Success:       true,
		OrderID:       order.ID,
		TransactionID: order.ID,
		Provider:      "paypal",
	}, nil
	*/

	// ========================================================
	// CURRENT: MOCK IMPLEMENTATION (for testing)
	// ========================================================
	// In production, comment out this section and uncomment provider above

	mockOrderID := fmt.Sprintf("ORDER_%s_%d", payReq.DriverID[:8], time.Now().Unix())
	mockTxnID := fmt.Sprintf("TXN_%s_%d", payReq.DriverID[:8], time.Now().UnixNano())

	logger.Info("✅ MOCK PAYMENT PROCESSED",
		"orderID", mockOrderID,
		"transactionID", mockTxnID,
		"amount", payReq.Amount,
		"note", "Replace with real payment provider")

	return &PaymentResult{
		Success:       true,
		OrderID:       mockOrderID,
		TransactionID: mockTxnID,
		Provider:      "mock", // Change to "razorpay", "stripe", or "paypal"
	}, nil
}

// ✅ GetWalletStatus returns detailed wallet and account status
func (s *service) GetWalletStatus(ctx context.Context, userID string) (*driverdto.WalletStatusResponse, error) {
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	// Get wallet through wallet service
	walletInfo, err := s.walletService.GetWallet(ctx, userID)
	if err != nil {
		// Return zero balance if wallet doesn't exist
		walletInfo = &walletdto.WalletResponse{
			Balance:     0,
			HeldBalance: 0,
			Currency:    "INR",
		}
	}

	// Calculate amount needed to lift restriction
	amountNeeded := 0.0
	if driver.IsRestricted {
		amountNeeded = (driver.MinBalanceThreshold * -1) - walletInfo.Balance
		if amountNeeded < 0 {
			amountNeeded = 0
		}
	}

	status := &driverdto.WalletStatusResponse{
		Balance:             walletInfo.Balance,
		HeldBalance:         walletInfo.HeldBalance,
		AvailableBalance:    walletInfo.Balance - walletInfo.HeldBalance,
		Currency:            walletInfo.Currency,
		IsRestricted:        driver.IsRestricted,
		AccountStatus:       driver.AccountStatus,
		RestrictionReason:   driver.RestrictionReason,
		MinBalanceThreshold: driver.MinBalanceThreshold,
		AmountNeededToLift:  amountNeeded,
		RestrictedAt:        driver.RestrictedAt,
		LastUpdated:         time.Now(),
	}

	return status, nil
}

// ✅ GetWalletTransactionHistory returns transaction history for driver
func (s *service) GetWalletTransactionHistory(ctx context.Context, userID string) ([]*driverdto.WalletTransactionResponse, error) {
	driver, err := s.repo.FindDriverByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	// For now, return mock transaction history
	// In production, fetch from wallet service or a dedicated transaction service
	logger.Info("fetching wallet transactions",
		"driverID", driver.ID,
		"userID", userID)

	// TODO: Implement proper transaction history from wallet service
	// This requires adding a method to wallet service Repository to fetch transactions by wallet ID
	
	return []*driverdto.WalletTransactionResponse{}, nil
}

// func (s *service) GetDriverStats(ctx context.Context, userID string) (*dto.DriverStatsResponse, error) {
// 	driver, err := s.repo.FindDriverByUserID(ctx, userID)
// 	if err != nil {
// 		return nil, response.NotFoundError("Driver profile")
// 	}

// 	return &driverdto.DriverStatsResponse{
// 		TotalTrips:       driver.TotalTrips,
// 		TotalEarnings:    driver.TotalEarnings,
// 		Rating:           driver.Rating,
// 		AcceptanceRate:   driver.AcceptanceRate,
// 		CancellationRate: driver.CancellationRate,
// 	}, nil
// }
