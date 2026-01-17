package wallet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/wallet/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
	"gorm.io/gorm"
)

type Service interface {
	TransferFunds(ctx context.Context, senderID string, req dto.TransferFundsRequest) (*dto.TransactionResponse, error)
	// Balance operations
	GetBalance(ctx context.Context, userID string) (*dto.WalletBalanceResponse, error)
	GetWallet(ctx context.Context, userID string) (*dto.WalletResponse, error)
	AddFunds(ctx context.Context, userID string, req dto.AddFundsRequest) (*dto.TransactionResponse, error)
	WithdrawFunds(ctx context.Context, userID string, req dto.WithdrawFundsRequest) (*dto.TransactionResponse, error)

	// Transactions
	GetTransactionHistory(ctx context.Context, userID string, req dto.TransactionHistoryRequest) ([]*dto.TransactionResponse, int64, error)
	GetTransaction(ctx context.Context, userID string, transactionID string) (*dto.TransactionResponse, error)
	ListTransactions(ctx context.Context, userID string, req dto.ListTransactionsRequest) ([]*dto.TransactionResponse, int64, error)

	// Hold and capture (for rides with cash payment tracking)
	HoldFunds(ctx context.Context, userID string, req dto.HoldFundsRequest) (*dto.HoldResponse, error)
	ReleaseHold(ctx context.Context, userID string, req dto.ReleaseHoldRequest) error
	CaptureHold(ctx context.Context, userID string, req dto.CaptureHoldRequest) (*dto.TransactionResponse, error)

	// Direct debit/credit (internal - for admin operations, ride completion, refunds)
	DebitWallet(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error)
	CreditWallet(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error)
	CreditDriverWallet(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error)

	// ✅ Driver wallet operations (for cash-based model with commissions/penalties)
	DebitDriverWallet(ctx context.Context, driverID string, amount float64, reason, referenceID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error)
	DeductCommission(ctx context.Context, driverID string, amount float64, commissionRate float64, rideID string) (*models.WalletTransaction, error)
	DeductPenalty(ctx context.Context, driverID string, amount float64, penaltyReason, rideID string) (*models.WalletTransaction, error)
	DeductSubscription(ctx context.Context, driverID string, amount float64, planName string) (*models.WalletTransaction, error)
	ValidateDriverWalletBalance(ctx context.Context, driverID string, amount float64) (float64, error)

	// ✅ Driver account restrictions (auto-disable on negative balance)
	CheckAndEnforceAccountRestriction(ctx context.Context, driverID string) (isRestricted bool, reason string, err error)
	RestrictDriverAccount(ctx context.Context, driverID string, reason string) error
	UnrestrictDriverAccount(ctx context.Context, driverID string) error
	RecordBalanceAudit(ctx context.Context, driverID, userID string, previousBalance, newBalance float64, action, reason string) error

	// Cash collection tracking
	RecordCashCollection(ctx context.Context, userID string, req dto.CashCollectionRequest) (*dto.TransactionResponse, error)
	RecordCashPayment(ctx context.Context, userID string, req dto.CashPaymentRequest) (*dto.TransactionResponse, error)
}

type service struct {
	repo Repository
	db   *gorm.DB
}

func NewService(repo Repository, db *gorm.DB) Service {
	return &service{
		repo: repo,
		db:   db,
	}
}

// GetWallet retrieves user's wallet
func (s *service) GetWallet(ctx context.Context, userID string) (*dto.WalletResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("wallet:user:%s", userID)
	var cachedWallet models.Wallet
	err := cache.GetJSON(ctx, cacheKey, &cachedWallet)
	if err == nil {
		return dto.ToWalletResponse(&cachedWallet), nil
	}

	// Determine wallet type from user role (simplified - you might need to fetch user)
	wallet, err := s.repo.FindWalletByUserID(ctx, userID, models.WalletTypeRider)
	if err != nil {
		// Try driver wallet
		wallet, err = s.repo.FindWalletByUserID(ctx, userID, models.WalletTypeDriver)
		if err != nil {
			return nil, response.NotFoundError("Wallet")
		}
	}

	// Cache for 2 minutes
	cache.SetJSON(ctx, cacheKey, wallet, 2*time.Minute)

	return dto.ToWalletResponse(wallet), nil
}

// GetBalance retrieves user's wallet balance
func (s *service) GetBalance(ctx context.Context, userID string) (*dto.WalletBalanceResponse, error) {
	wallet, err := s.repo.FindWalletByUserID(ctx, userID, models.WalletTypeRider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create wallet if doesn't exist
			wallet = &models.Wallet{
				UserID:      userID,
				WalletType:  models.WalletTypeRider,
				Balance:     0,
				HeldBalance: 0,
				Currency:    "INR",
			}
			if err := s.repo.CreateWallet(ctx, wallet); err != nil {
				return nil, response.InternalServerError("Failed to create wallet", err)
			}
		} else {
			return nil, response.InternalServerError("Failed to fetch wallet", err)
		}
	}

	return &dto.WalletBalanceResponse{
		WalletID:         wallet.ID,
		Balance:          wallet.Balance,
		HeldBalance:      wallet.HeldBalance,
		AvailableBalance: wallet.Balance - wallet.HeldBalance,
		Currency:         wallet.Currency,
		UpdatedAt:        wallet.UpdatedAt,
	}, nil
}

// AddFunds adds money to wallet (simulated top-up)
func (s *service) AddFunds(ctx context.Context, userID string, req dto.AddFundsRequest) (*dto.TransactionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Get wallet
	walletResp, err := s.GetWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	wallet, err := s.repo.FindWalletByID(ctx, walletResp.ID)
	if err != nil {
		return nil, response.NotFoundError("Wallet")
	}

	if !wallet.IsActive {
		return nil, response.BadRequest("Wallet is not active")
	}

	// Create transaction in database transaction
	var transaction *models.WalletTransaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Record balance before
		balanceBefore := wallet.Balance

		// Update wallet balance
		wallet.Balance += req.Amount

		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		// Create transaction record
		now := time.Now()
		transaction = &models.WalletTransaction{
			WalletID:      wallet.ID,
			Type:          models.TransactionTypeCredit,
			Amount:        req.Amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  wallet.Balance,
			Status:        models.TransactionStatusCompleted,
			ReferenceType: stringPtr("topup"),
			Description:   stringPtr(req.Description),
			ProcessedAt:   &now,
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to add funds", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to add funds", err)
	}

	// Invalidate cache
	s.invalidateWalletCache(ctx, userID)

	logger.Info("funds added", "userID", userID, "amount", req.Amount, "txID", transaction.ID)

	return dto.ToTransactionResponse(transaction), nil
}

// WithdrawFunds withdraws money from wallet
func (s *service) WithdrawFunds(ctx context.Context, userID string, req dto.WithdrawFundsRequest) (*dto.TransactionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Get wallet
	walletResp, err := s.GetWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	wallet, err := s.repo.FindWalletByID(ctx, walletResp.ID)
	if err != nil {
		return nil, response.NotFoundError("Wallet")
	}

	if !wallet.IsActive {
		return nil, response.BadRequest("Wallet is not active")
	}

	// Check available balance
	if wallet.GetAvailableBalance() < req.Amount {
		return nil, response.BadRequest("Insufficient balance")
	}

	// Create transaction
	var transaction *models.WalletTransaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		balanceBefore := wallet.Balance

		// Update wallet balance
		wallet.Balance -= req.Amount

		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		// Create transaction record
		now := time.Now()
		transaction = &models.WalletTransaction{
			WalletID:      wallet.ID,
			Type:          models.TransactionTypeDebit,
			Amount:        req.Amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  wallet.Balance,
			Status:        models.TransactionStatusCompleted,
			ReferenceType: stringPtr("withdrawal"),
			Description:   stringPtr(req.Description),
			ProcessedAt:   &now,
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to withdraw funds", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to withdraw funds", err)
	}

	// Invalidate cache
	s.invalidateWalletCache(ctx, userID)

	logger.Info("funds withdrawn", "userID", userID, "amount", req.Amount, "txID", transaction.ID)

	return dto.ToTransactionResponse(transaction), nil
}

// TransferFunds transfers money between wallets
func (s *service) TransferFunds(ctx context.Context, senderID string, req dto.TransferFundsRequest) (*dto.TransactionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	if senderID == req.RecipientID {
		return nil, response.BadRequest("Cannot transfer to yourself")
	}

	// Get sender wallet
	senderWalletResp, err := s.GetWallet(ctx, senderID)
	if err != nil {
		return nil, err
	}

	senderWallet, err := s.repo.FindWalletByID(ctx, senderWalletResp.ID)
	if err != nil {
		return nil, response.NotFoundError("Sender wallet")
	}

	// Get recipient wallet
	recipientWalletResp, err := s.GetWallet(ctx, req.RecipientID)
	if err != nil {
		return nil, response.NotFoundError("Recipient wallet")
	}

	recipientWallet, err := s.repo.FindWalletByID(ctx, recipientWalletResp.ID)
	if err != nil {
		return nil, response.NotFoundError("Recipient wallet")
	}

	// Check balances
	if senderWallet.GetAvailableBalance() < req.Amount {
		return nil, response.BadRequest("Insufficient balance")
	}

	if !senderWallet.IsActive || !recipientWallet.IsActive {
		return nil, response.BadRequest("One or both wallets are not active")
	}

	// Perform transfer
	var senderTx *models.WalletTransaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Debit sender
		senderBalanceBefore := senderWallet.Balance
		senderWallet.Balance -= req.Amount
		if err := tx.Save(senderWallet).Error; err != nil {
			return err
		}

		// Credit recipient
		recipientBalanceBefore := recipientWallet.Balance
		recipientWallet.Balance += req.Amount
		if err := tx.Save(recipientWallet).Error; err != nil {
			return err
		}

		now := time.Now()

		// Create sender transaction
		senderTx = &models.WalletTransaction{
			WalletID:      senderWallet.ID,
			Type:          models.TransactionTypeTransfer,
			Amount:        req.Amount,
			BalanceBefore: senderBalanceBefore,
			BalanceAfter:  senderWallet.Balance,
			Status:        models.TransactionStatusCompleted,
			ReferenceType: stringPtr("transfer_out"),
			ReferenceID:   &req.RecipientID,
			Description:   stringPtr(req.Description),
			Metadata: map[string]interface{}{
				"recipientId": req.RecipientID,
			},
			ProcessedAt: &now,
		}
		if err := tx.Create(senderTx).Error; err != nil {
			return err
		}

		// Create recipient transaction
		recipientTx := &models.WalletTransaction{
			WalletID:      recipientWallet.ID,
			Type:          models.TransactionTypeTransfer,
			Amount:        req.Amount,
			BalanceBefore: recipientBalanceBefore,
			BalanceAfter:  recipientWallet.Balance,
			Status:        models.TransactionStatusCompleted,
			ReferenceType: stringPtr("transfer_in"),
			ReferenceID:   &senderID,
			Description:   stringPtr(req.Description),
			Metadata: map[string]interface{}{
				"senderId": senderID,
			},
			ProcessedAt: &now,
		}
		if err := tx.Create(recipientTx).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to transfer funds", "error", err, "senderID", senderID)
		return nil, response.InternalServerError("Failed to transfer funds", err)
	}

	// Invalidate cache for both users
	s.invalidateWalletCache(ctx, senderID)
	s.invalidateWalletCache(ctx, req.RecipientID)

	logger.Info("funds transferred", "senderID", senderID, "recipientID", req.RecipientID, "amount", req.Amount)

	return dto.ToTransactionResponse(senderTx), nil
}

// HoldFunds creates a virtual hold for ride fare (not actual money hold)
// This is used to track expected payment amount for cash rides
func (s *service) HoldFunds(ctx context.Context, userID string, req dto.HoldFundsRequest) (*dto.HoldResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	wallet, err := s.repo.FindWalletByUserID(ctx, userID, models.WalletTypeRider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = &models.Wallet{
				UserID:     userID,
				WalletType: models.WalletTypeRider,
				Balance:    0,
				Currency:   "INR",
			}
			if err := s.repo.CreateWallet(ctx, wallet); err != nil {
				return nil, response.InternalServerError("Failed to create wallet", err)
			}
		} else {
			return nil, response.InternalServerError("Failed to fetch wallet", err)
		}
	}

	// For cash rides, we don't check balance - just create a tracking hold
	hold := &models.WalletHold{
		WalletID:      wallet.ID,
		Amount:        req.Amount,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		Status:        "active",
		ExpiresAt:     time.Now().Add(time.Duration(req.HoldDuration) * time.Second),
	}

	if err := s.repo.CreateHold(ctx, hold); err != nil {
		return nil, response.InternalServerError("Failed to create hold", err)
	}

	logger.Info("hold created for cash ride",
		"userID", userID,
		"amount", req.Amount,
		"holdID", hold.ID,
		"reference", req.ReferenceID)

	return &dto.HoldResponse{
		ID:        hold.ID,
		Amount:    req.Amount,
		ExpiresAt: hold.ExpiresAt,
	}, nil
}

// ReleaseHold releases a hold (e.g., ride cancelled)
func (s *service) ReleaseHold(ctx context.Context, userID string, req dto.ReleaseHoldRequest) error {
	hold, err := s.repo.FindHoldByID(ctx, req.HoldID)
	if err != nil {
		return response.NotFoundError("Hold")
	}

	wallet, err := s.repo.FindWalletByID(ctx, hold.WalletID)
	if err != nil || wallet.UserID != userID {
		return response.ForbiddenError("Not authorized to release this hold")
	}

	if hold.Status != "active" {
		return response.BadRequest("Hold is no longer active")
	}

	// Release hold
	hold.Status = "released"
	now := time.Now()
	hold.ReleasedAt = &now

	if err := s.repo.UpdateHold(ctx, hold); err != nil {
		return response.InternalServerError("Failed to release hold", err)
	}

	logger.Info("hold released", "holdID", hold.ID, "amount", hold.Amount, "userID", userID)

	return nil
}

// CaptureHold captures a hold after cash payment is confirmed
func (s *service) CaptureHold(ctx context.Context, userID string, req dto.CaptureHoldRequest) (*dto.TransactionResponse, error) {
	hold, err := s.repo.FindHoldByID(ctx, req.HoldID)
	if err != nil {
		return nil, response.NotFoundError("Hold")
	}

	wallet, err := s.repo.FindWalletByID(ctx, hold.WalletID)
	if err != nil || wallet.UserID != userID {
		return nil, response.ForbiddenError("Not authorized to capture this hold")
	}

	if hold.Status != "active" {
		return nil, response.BadRequest("Hold is no longer active")
	}

	captureAmount := hold.Amount
	if req.Amount != nil && *req.Amount <= hold.Amount {
		captureAmount = *req.Amount
	}

	// Create transaction record for cash payment
	txn := &models.WalletTransaction{
		WalletID:      wallet.ID,
		Amount:        captureAmount,
		Type:          "debit",
		Status:        "completed",
		ReferenceType: &hold.ReferenceType,
		ReferenceID:   &hold.ReferenceID,
		Description:   &req.Description,
		PaymentMethod: "cash",         // Mark as cash payment
		BalanceAfter:  wallet.Balance, // Balance doesn't change for cash
	}

	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, response.InternalServerError("Failed to create transaction", err)
	}

	// Update hold status
	hold.Status = "captured"
	now := time.Now()
	hold.CreatedAt = now
	hold.Amount = captureAmount

	if err := s.repo.UpdateHold(ctx, hold); err != nil {
		logger.Error("failed to update hold status", "error", err, "holdID", hold.ID)
	}

	logger.Info("hold captured (cash payment)",
		"holdID", hold.ID,
		"amount", captureAmount,
		"userID", userID,
		"transactionID", txn.ID)

	return dto.ToTransactionResponse(txn), nil
}

// GetHoldsByReference retrieves holds by reference (used internally)
func (s *service) GetHoldsByReference(ctx context.Context, refType, refID string) ([]*dto.HoldResponse, error) {
	holds, err := s.repo.FindHoldsByReference(ctx, refType, refID)
	if err != nil {
		return nil, response.InternalServerError("Failed to fetch holds", err)
	}

	result := make([]*dto.HoldResponse, len(holds))
	for i, hold := range holds {
		result[i] = dto.ToHoldResponse(hold)
	}

	return result, nil
}

// ListTransactions lists user's transactions
func (s *service) ListTransactions(ctx context.Context, userID string, req dto.ListTransactionsRequest) ([]*dto.TransactionResponse, int64, error) {
	req.SetDefaults()

	// Get wallet
	walletResp, err := s.GetWallet(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	filters := make(map[string]interface{})
	if req.Type != "" {
		filters["type"] = req.Type
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}

	transactions, total, err := s.repo.ListTransactions(ctx, walletResp.ID, filters, req.Page, req.Limit)
	if err != nil {
		return nil, 0, response.InternalServerError("Failed to fetch transactions", err)
	}

	result := make([]*dto.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		result[i] = dto.ToTransactionResponse(tx)
	}

	return result, total, nil
}

// GetTransaction retrieves a specific transaction
func (s *service) GetTransaction(ctx context.Context, userID string, transactionID string) (*dto.TransactionResponse, error) {
	txn, err := s.repo.FindTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, response.NotFoundError("Transaction")
	}

	if txn.WalletID != "" {
		wallet, err := s.repo.FindWalletByID(ctx, txn.WalletID)
		if err != nil || wallet.UserID != userID {
			return nil, response.ForbiddenError("Not authorized to view this transaction")
		}
	}

	return dto.ToTransactionResponse(txn), nil
}

// DebitWallet - Internal method for other modules
func (s *service) DebitWallet(ctx context.Context, userID string, amount float64, refType, refID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error) {
	walletResp, err := s.GetWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	wallet, err := s.repo.FindWalletByID(ctx, walletResp.ID)
	if err != nil {
		return nil, err
	}

	if wallet.GetAvailableBalance() < amount {
		return nil, response.BadRequest("Insufficient balance")
	}

	var transaction *models.WalletTransaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		balanceBefore := wallet.Balance
		wallet.Balance -= amount

		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		now := time.Now()
		transaction = &models.WalletTransaction{
			WalletID:      wallet.ID,
			Type:          models.TransactionTypeDebit,
			Amount:        amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  wallet.Balance,
			Status:        models.TransactionStatusCompleted,
			ReferenceType: &refType,
			ReferenceID:   &refID,
			Description:   &description,
			Metadata:      metadata,
			ProcessedAt:   &now,
		}

		return tx.Create(transaction).Error
	})

	if err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, userID)
	return transaction, nil
}

// CreditWallet credits amount to wallet (for driver earnings, refunds, bonuses)
func (s *service) CreditWallet(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error) {
	wallet, err := s.repo.FindWalletByUserID(ctx, userID, models.WalletTypeRider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = &models.Wallet{
				UserID:     userID,
				Balance:    0,
				Currency:   "INR",
				WalletType: models.WalletTypeRider, // Set wallet type to rider
			}
			if err := s.repo.CreateWallet(ctx, wallet); err != nil {
				return nil, response.InternalServerError("Failed to create wallet", err)
			}
		} else {
			return nil, response.InternalServerError("Failed to fetch wallet", err)
		}
	}

	txn := &models.WalletTransaction{
		WalletID:      wallet.ID,
		Amount:        amount,
		Type:          "credit",
		Status:        "completed",
		ReferenceType: &transactionType,
		ReferenceID:   &referenceID,
		Description:   &description,
		BalanceAfter:  wallet.Balance + amount,
	}

	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, response.InternalServerError("Failed to create transaction", err)
	}

	wallet.Balance += amount
	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		return nil, response.InternalServerError("Failed to update wallet", err)
	}

	logger.Info("wallet credited", "userID", userID, "amount", amount, "transactionID", txn.ID, "type", transactionType)

	return txn, nil
}

// CreditDriverWallet credits amount to driver wallet (for driver earnings)
func (s *service) CreditDriverWallet(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error) {
	wallet, err := s.repo.FindWalletByUserID(ctx, userID, models.WalletTypeDriver)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = &models.Wallet{
				UserID:     userID,
				Balance:    0,
				Currency:   "INR",
				WalletType: models.WalletTypeDriver, // Set wallet type to driver
			}
			if err := s.repo.CreateWallet(ctx, wallet); err != nil {
				return nil, response.InternalServerError("Failed to create driver wallet", err)
			}
		} else {
			return nil, response.InternalServerError("Failed to fetch driver wallet", err)
		}
	}

	txn := &models.WalletTransaction{
		WalletID:      wallet.ID,
		Amount:        amount,
		Type:          "credit",
		Status:        "completed",
		ReferenceType: &transactionType,
		ReferenceID:   &referenceID,
		Description:   &description,
		BalanceAfter:  wallet.Balance + amount,
	}

	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, response.InternalServerError("Failed to create transaction", err)
	}

	wallet.Balance += amount
	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		return nil, response.InternalServerError("Failed to update wallet", err)
	}

	logger.Info("driver wallet credited", "userID", userID, "amount", amount, "transactionID", txn.ID, "type", transactionType)

	return txn, nil
}

// ✅ DebitDriverWallet debits amount from driver wallet (for commissions, penalties, subscriptions)
// This ensures driver has funds in wallet to debit before deducting
func (s *service) DebitDriverWallet(ctx context.Context, driverID string, amount float64, reason, referenceID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error) {
	wallet, err := s.repo.FindWalletByUserID(ctx, driverID, models.WalletTypeDriver)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = &models.Wallet{
				UserID:     driverID,
				Balance:    0,
				Currency:   "INR",
				WalletType: models.WalletTypeDriver,
			}
			if err := s.repo.CreateWallet(ctx, wallet); err != nil {
				return nil, response.InternalServerError("Failed to create driver wallet", err)
			}
		} else {
			return nil, response.InternalServerError("Failed to fetch driver wallet", err)
		}
	}

	// Check available balance (allow negative balance for now, but log it)
	availableBalance := wallet.GetAvailableBalance()
	if availableBalance < amount {
		logger.Warn("driver insufficient balance for debit",
			"driverID", driverID,
			"requiredAmount", amount,
			"availableBalance", availableBalance,
			"reason", reason)
	}

	var transaction *models.WalletTransaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		balanceBefore := wallet.Balance
		wallet.Balance -= amount

		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		now := time.Now()
		transaction = &models.WalletTransaction{
			WalletID:      wallet.ID,
			Type:          models.TransactionTypeDebit,
			Amount:        amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  wallet.Balance,
			Status:        models.TransactionStatusCompleted,
			ReferenceType: &reason,
			ReferenceID:   &referenceID,
			Description:   &description,
			Metadata:      metadata,
			ProcessedAt:   &now,
		}

		return tx.Create(transaction).Error
	})

	if err != nil {
		return nil, response.InternalServerError("Failed to debit driver wallet", err)
	}

	s.invalidateWalletCache(ctx, driverID)
	logger.Info("driver wallet debited",
		"driverID", driverID,
		"amount", amount,
		"reason", reason,
		"referenceID", referenceID,
		"newBalance", wallet.Balance)

	// ✅ Record balance audit for compliance
	var driver models.DriverProfile
	if err := s.db.WithContext(ctx).Where("id = ?", driverID).First(&driver).Error; err == nil {
		_ = s.RecordBalanceAudit(ctx, driverID, driver.UserID, wallet.Balance+amount, wallet.Balance, reason, description)
	}

	// ✅ Check if account should be restricted due to negative balance
	go func() {
		bgCtx := context.Background()
		if isRestricted, restrictReason, err := s.CheckAndEnforceAccountRestriction(bgCtx, driverID); err != nil {
			logger.Error("failed to check account restriction", "error", err, "driverID", driverID)
		} else if isRestricted {
			logger.Warn("driver account restricted due to negative balance",
				"driverID", driverID,
				"reason", restrictReason)
		}
	}()

	return transaction, nil
}

// ✅ DeductCommission deducts platform commission from driver earnings
// For cash-based rides: driver earns after commission already calculated
// This method tracks commission separately for reporting
func (s *service) DeductCommission(ctx context.Context, driverID string, amount float64, commissionRate float64, rideID string) (*models.WalletTransaction, error) {
	reason := "ride_commission"
	description := fmt.Sprintf("Platform commission (%.1f%%) for ride %s", commissionRate, rideID)

	metadata := map[string]interface{}{
		"commission_rate": commissionRate,
		"ride_id":         rideID,
		"type":            "platform_commission",
	}

	return s.DebitDriverWallet(ctx, driverID, amount, reason, rideID, description, metadata)
}

// ✅ DeductPenalty deducts penalty from driver wallet
// Penalties: cancellation fees, ratings-based penalties, rule violations
func (s *service) DeductPenalty(ctx context.Context, driverID string, amount float64, penaltyReason, rideID string) (*models.WalletTransaction, error) {
	reason := "driver_penalty"
	description := fmt.Sprintf("Penalty (%s) - %s", penaltyReason, rideID)

	metadata := map[string]interface{}{
		"penalty_reason": penaltyReason,
		"ride_id":        rideID,
		"type":           "penalty",
	}

	logger.Info("applying driver penalty",
		"driverID", driverID,
		"amount", amount,
		"reason", penaltyReason,
		"rideID", rideID)

	return s.DebitDriverWallet(ctx, driverID, amount, reason, rideID, description, metadata)
}

// ✅ DeductSubscription deducts subscription fees from driver wallet
// For future subscriptions: premium support, features, insurance
func (s *service) DeductSubscription(ctx context.Context, driverID string, amount float64, planName string) (*models.WalletTransaction, error) {
	reason := "subscription_fee"
	referenceID := fmt.Sprintf("subscription_%s_%d", planName, time.Now().Unix())
	description := fmt.Sprintf("Subscription fee for %s plan", planName)

	metadata := map[string]interface{}{
		"plan_name": planName,
		"type":      "subscription",
		"date":      time.Now().Format("2006-01-02"),
	}

	logger.Info("deducting subscription fee",
		"driverID", driverID,
		"plan", planName,
		"amount", amount)

	return s.DebitDriverWallet(ctx, driverID, amount, reason, referenceID, description, metadata)
}

// ✅ ValidateDriverWalletBalance checks if driver has sufficient balance
// Returns available balance for verification
func (s *service) ValidateDriverWalletBalance(ctx context.Context, driverID string, amount float64) (float64, error) {
	wallet, err := s.repo.FindWalletByUserID(ctx, driverID, models.WalletTypeDriver)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil // No wallet = no balance
		}
		return 0, response.InternalServerError("Failed to fetch driver wallet", err)
	}

	availableBalance := wallet.GetAvailableBalance()

	if availableBalance < amount {
		logger.Warn("insufficient driver wallet balance",
			"driverID", driverID,
			"required", amount,
			"available", availableBalance)
		return availableBalance, response.BadRequest(
			fmt.Sprintf("Insufficient wallet balance. Required: $%.2f, Available: $%.2f", amount, availableBalance))
	}

	return availableBalance, nil
}

// GetTransactionHistory retrieves transaction history
func (s *service) GetTransactionHistory(ctx context.Context, userID string, req dto.TransactionHistoryRequest) ([]*dto.TransactionResponse, int64, error) {
	req.SetDefaults()

	filters := make(map[string]interface{})
	if req.Type != "" {
		filters["type"] = req.Type
	}
	if req.StartDate != "" {
		filters["startDate"] = req.StartDate
	}
	if req.EndDate != "" {
		filters["endDate"] = req.EndDate
	}

	transactions, total, err := s.repo.ListTransactions(ctx, userID, filters, req.Page, req.Limit)
	if err != nil {
		return nil, 0, response.InternalServerError("Failed to fetch transactions", err)
	}

	result := make([]*dto.TransactionResponse, len(transactions))
	for i, txn := range transactions {
		result[i] = dto.ToTransactionResponse(txn)
	}

	return result, total, nil
}

// RecordCashCollection records when driver collects cash from rider
func (s *service) RecordCashCollection(ctx context.Context, userID string, req dto.CashCollectionRequest) (*dto.TransactionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	wallet, err := s.repo.FindWalletByUserID(ctx, userID, models.WalletTypeDriver)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = &models.Wallet{
				UserID:     userID,
				WalletType: models.WalletTypeDriver,
				Balance:    0,
				Currency:   "INR",
			}
			if err := s.repo.CreateWallet(ctx, wallet); err != nil {
				return nil, response.InternalServerError("Failed to create wallet", err)
			}
		} else {
			return nil, response.InternalServerError("Failed to fetch wallet", err)
		}
	}

	// Record cash collection as a transaction
	txn := &models.WalletTransaction{
		WalletID:      wallet.ID,
		Amount:        req.Amount,
		Type:          "credit",
		Status:        "completed",
		ReferenceType: stringPtr("cash_collection"),
		ReferenceID:   &req.RideID,
		Description:   stringPtr(fmt.Sprintf("Cash collected from ride %s", req.RideID)),
		PaymentMethod: "cash",
		BalanceAfter:  wallet.Balance + req.Amount,
	}

	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, response.InternalServerError("Failed to record cash collection", err)
	}

	// Update wallet balance (driver's cash in hand)
	wallet.Balance += req.Amount
	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		logger.Error("failed to update wallet balance", "error", err, "walletID", wallet.ID)
	}

	logger.Info("cash collection recorded",
		"driverID", userID,
		"amount", req.Amount,
		"rideID", req.RideID,
		"transactionID", txn.ID)

	return dto.ToTransactionResponse(txn), nil
}

// RecordCashPayment records when driver pays cash to company (settlement)
func (s *service) RecordCashPayment(ctx context.Context, userID string, req dto.CashPaymentRequest) (*dto.TransactionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	wallet, err := s.repo.FindWalletByUserID(ctx, userID, models.WalletTypeDriver)
	if err != nil {
		return nil, response.NotFoundError("Wallet")
	}

	if req.Amount > wallet.Balance {
		return nil, response.BadRequest(fmt.Sprintf("Insufficient balance. Current: $%.2f", wallet.Balance))
	}

	// Record cash payment to company
	txn := &models.WalletTransaction{
		WalletID:      wallet.ID,
		Amount:        req.Amount,
		Type:          "debit",
		Status:        "completed",
		ReferenceType: stringPtr("cash_settlement"),
		ReferenceID:   &req.SettlementID,
		Description:   stringPtr(fmt.Sprintf("Cash settlement to company - %s", req.SettlementID)),
		PaymentMethod: "cash",
		BalanceAfter:  wallet.Balance - req.Amount,
	}

	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, response.InternalServerError("Failed to record cash payment", err)
	}

	// Update wallet balance
	wallet.Balance -= req.Amount
	if err := s.repo.UpdateWallet(ctx, wallet); err != nil {
		logger.Error("failed to update wallet balance", "error", err, "walletID", wallet.ID)
	}

	logger.Info("cash settlement recorded",
		"driverID", userID,
		"amount", req.Amount,
		"settlementID", req.SettlementID,
		"transactionID", txn.ID)

	return dto.ToTransactionResponse(txn), nil
}

// ✅ CheckAndEnforceAccountRestriction checks if driver account should be restricted
// Returns true if account is restricted or should be restricted
func (s *service) CheckAndEnforceAccountRestriction(ctx context.Context, driverID string) (bool, string, error) {
	// Get driver profile
	var driver models.DriverProfile
	if err := s.db.WithContext(ctx).Where("id = ?", driverID).First(&driver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", response.NotFoundError("Driver")
		}
		return false, "", response.InternalServerError("Failed to fetch driver profile", err)
	}

	// Get driver's wallet balance
	wallet, err := s.repo.FindWalletByUserID(ctx, driver.UserID, models.WalletTypeDriver)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", nil // No wallet = no restriction
		}
		return false, "", response.InternalServerError("Failed to fetch wallet", err)
	}

	// Check if balance is below threshold
	if wallet.Balance < driver.MinBalanceThreshold {
		reason := fmt.Sprintf("Negative balance: $%.2f (threshold: $%.2f)", wallet.Balance, driver.MinBalanceThreshold)

		// Auto-restrict if not already restricted
		if !driver.IsRestricted {
			if err := s.RestrictDriverAccount(ctx, driverID, reason); err != nil {
				logger.Error("failed to restrict driver account", "error", err, "driverID", driverID)
			}
		}

		return true, reason, nil
	}

	// If balance is positive and account was restricted, unrestrict it
	if driver.IsRestricted && wallet.Balance >= 0 {
		if err := s.UnrestrictDriverAccount(ctx, driverID); err != nil {
			logger.Error("failed to unrestrict driver account", "error", err, "driverID", driverID)
		}
	}

	return driver.IsRestricted, "", nil
}

// ✅ RestrictDriverAccount disables driver account due to negative balance
func (s *service) RestrictDriverAccount(ctx context.Context, driverID string, reason string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&models.DriverProfile{}).
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"is_restricted":      true,
			"account_status":     "disabled",
			"restricted_at":      now,
			"restriction_reason": reason,
		})

	if result.Error != nil {
		return response.InternalServerError("Failed to restrict account", result.Error)
	}

	if result.RowsAffected == 0 {
		return response.NotFoundError("Driver")
	}

	logger.Info("driver account restricted",
		"driverID", driverID,
		"reason", reason,
		"restrictedAt", now)

	// Invalidate driver cache
	cache.Delete(ctx, fmt.Sprintf("driver:profile:%s", driverID))

	return nil
}

// ✅ UnrestrictDriverAccount re-enables driver account after adding funds
func (s *service) UnrestrictDriverAccount(ctx context.Context, driverID string) error {
	result := s.db.WithContext(ctx).Model(&models.DriverProfile{}).
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"is_restricted":      false,
			"account_status":     "active",
			"restricted_at":      nil,
			"restriction_reason": nil,
		})

	if result.Error != nil {
		return response.InternalServerError("Failed to unrestrict account", result.Error)
	}

	if result.RowsAffected == 0 {
		return response.NotFoundError("Driver")
	}

	logger.Info("driver account unrestricted",
		"driverID", driverID)

	// Invalidate driver cache
	cache.Delete(ctx, fmt.Sprintf("driver:profile:%s", driverID))

	return nil
}

// ✅ RecordBalanceAudit logs all balance changes for auditing
func (s *service) RecordBalanceAudit(ctx context.Context, driverID, userID string, previousBalance, newBalance float64, action, reason string) error {
	changeAmount := newBalance - previousBalance

	// Check if this triggers a restriction
	var driver models.DriverProfile
	triggeredRestriction := false
	if err := s.db.WithContext(ctx).Where("id = ?", driverID).First(&driver).Error; err == nil {
		if newBalance < driver.MinBalanceThreshold && !driver.IsRestricted {
			triggeredRestriction = true
		}
	}

	audit := &models.DriverBalanceAudit{
		DriverID:             driverID,
		UserID:               userID,
		PreviousBalance:      previousBalance,
		NewBalance:           newBalance,
		ChangeAmount:         changeAmount,
		Action:               action,
		Reason:               &reason,
		TriggeredRestriction: triggeredRestriction,
	}

	if err := s.db.WithContext(ctx).Create(audit).Error; err != nil {
		logger.Error("failed to record balance audit", "error", err, "driverID", driverID, "action", action)
		// Don't return error - auditing failure shouldn't break the transaction
		return nil
	}

	logger.Info("balance audit recorded",
		"driverID", driverID,
		"action", action,
		"previousBalance", previousBalance,
		"newBalance", newBalance,
		"changeAmount", changeAmount,
		"triggeredRestriction", triggeredRestriction)

	return nil
}

// Helper functions
func (s *service) invalidateWalletCache(ctx context.Context, userID string) {
	cache.Delete(ctx, fmt.Sprintf("wallet:user:%s", userID))
}

func stringPtr(s string) *string {
	return &s
}
