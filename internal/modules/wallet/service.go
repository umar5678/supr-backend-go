package wallet

import (
	"context"
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
	// Wallet operations
	GetWallet(ctx context.Context, userID string) (*dto.WalletResponse, error)
	GetBalance(ctx context.Context, userID string) (float64, error)

	// Transaction operations
	AddFunds(ctx context.Context, userID string, req dto.AddFundsRequest) (*dto.TransactionResponse, error)
	WithdrawFunds(ctx context.Context, userID string, req dto.WithdrawFundsRequest) (*dto.TransactionResponse, error)
	TransferFunds(ctx context.Context, senderID string, req dto.TransferFundsRequest) (*dto.TransactionResponse, error)
	ListTransactions(ctx context.Context, userID string, req dto.ListTransactionsRequest) ([]*dto.TransactionResponse, int64, error)
	GetTransaction(ctx context.Context, userID, txID string) (*dto.TransactionResponse, error)

	// Hold operations (for rides, orders, etc.)
	HoldFunds(ctx context.Context, userID string, req dto.HoldFundsRequest) (*dto.HoldResponse, error)
	ReleaseHold(ctx context.Context, userID string, req dto.ReleaseHoldRequest) error
	CaptureHold(ctx context.Context, userID string, req dto.CaptureHoldRequest) (*dto.TransactionResponse, error)
	GetHoldsByReference(ctx context.Context, refType, refID string) ([]*dto.HoldResponse, error)

	// Internal operations (used by other modules)
	DebitWallet(ctx context.Context, userID string, amount float64, refType, refID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error)
	CreditWallet(ctx context.Context, userID string, amount float64, refType, refID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error)
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

// GetBalance retrieves available balance
func (s *service) GetBalance(ctx context.Context, userID string) (float64, error) {
	wallet, err := s.GetWallet(ctx, userID)
	if err != nil {
		return 0, err
	}
	return wallet.AvailableBalance, nil
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

// HoldFunds places a hold on funds (for pending transactions)
func (s *service) HoldFunds(ctx context.Context, userID string, req dto.HoldFundsRequest) (*dto.HoldResponse, error) {
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

	// Check available balance
	if wallet.GetAvailableBalance() < req.Amount {
		return nil, response.BadRequest("Insufficient balance")
	}

	// Create hold
	var hold *models.WalletHold
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update wallet held balance
		wallet.HeldBalance += req.Amount
		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		// Create hold record
		expiresAt := time.Now().Add(time.Duration(req.HoldDuration) * time.Minute)
		hold = &models.WalletHold{
			WalletID:      wallet.ID,
			Amount:        req.Amount,
			ReferenceType: req.ReferenceType,
			ReferenceID:   req.ReferenceID,
			Status:        models.TransactionStatusHeld,
			ExpiresAt:     expiresAt,
		}

		if err := tx.Create(hold).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to hold funds", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to hold funds", err)
	}

	// Invalidate cache
	s.invalidateWalletCache(ctx, userID)

	logger.Info("funds held", "userID", userID, "amount", req.Amount, "holdID", hold.ID)

	return dto.ToHoldResponse(hold), nil
}

// ReleaseHold releases a hold without capturing
func (s *service) ReleaseHold(ctx context.Context, userID string, req dto.ReleaseHoldRequest) error {
	hold, err := s.repo.FindHoldByID(ctx, req.HoldID)
	if err != nil {
		return response.NotFoundError("Hold")
	}

	wallet, err := s.repo.FindWalletByID(ctx, hold.WalletID)
	if err != nil {
		return response.NotFoundError("Wallet")
	}

	// Verify ownership
	if wallet.UserID != userID {
		return response.ForbiddenError("Not authorized to release this hold")
	}

	if hold.Status != models.TransactionStatusHeld {
		return response.BadRequest("Hold is not in held status")
	}

	// Release hold
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update wallet
		wallet.HeldBalance -= hold.Amount
		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		// Update hold
		now := time.Now()
		hold.Status = models.TransactionStatusReleased
		hold.ReleasedAt = &now
		if err := tx.Save(hold).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to release hold", "error", err, "holdID", req.HoldID)
		return response.InternalServerError("Failed to release hold", err)
	}

	// Invalidate cache
	s.invalidateWalletCache(ctx, userID)

	logger.Info("hold released", "userID", userID, "holdID", req.HoldID)

	return nil
}

// CaptureHold captures a hold and creates a transaction
func (s *service) CaptureHold(ctx context.Context, userID string, req dto.CaptureHoldRequest) (*dto.TransactionResponse, error) {
	hold, err := s.repo.FindHoldByID(ctx, req.HoldID)
	if err != nil {
		return nil, response.NotFoundError("Hold")
	}

	wallet, err := s.repo.FindWalletByID(ctx, hold.WalletID)
	if err != nil {
		return nil, response.NotFoundError("Wallet")
	}

	// Verify ownership
	if wallet.UserID != userID {
		return nil, response.ForbiddenError("Not authorized to capture this hold")
	}

	if hold.Status != models.TransactionStatusHeld {
		return nil, response.BadRequest("Hold is not in held status")
	}

	// Determine capture amount
	captureAmount := hold.Amount
	if req.Amount != nil {
		if *req.Amount > hold.Amount {
			return nil, response.BadRequest("Capture amount exceeds hold amount")
		}
		captureAmount = *req.Amount
	}

	// Capture hold
	var transaction *models.WalletTransaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		balanceBefore := wallet.Balance

		// Deduct from balance and held balance
		wallet.Balance -= captureAmount
		wallet.HeldBalance -= hold.Amount // Release full hold amount

		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		// Update hold
		now := time.Now()
		hold.Status = models.TransactionStatusReleased
		hold.ReleasedAt = &now
		if err := tx.Save(hold).Error; err != nil {
			return err
		}

		// Create transaction
		description := req.Description
		if description == "" {
			description = fmt.Sprintf("Captured from hold %s", hold.ID)
		}

		transaction = &models.WalletTransaction{
			WalletID:      wallet.ID,
			Type:          models.TransactionTypeDebit,
			Amount:        captureAmount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  wallet.Balance,
			Status:        models.TransactionStatusCompleted,
			ReferenceType: &hold.ReferenceType,
			ReferenceID:   &hold.ReferenceID,
			Description:   &description,
			Metadata: map[string]interface{}{
				"holdId":         hold.ID,
				"heldAmount":     hold.Amount,
				"capturedAmount": captureAmount,
			},
			ProcessedAt: &now,
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// If partial capture, release remaining
		if captureAmount < hold.Amount {
			wallet.Balance += (hold.Amount - captureAmount)
			if err := tx.Save(wallet).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to capture hold", "error", err, "holdID", req.HoldID)
		return nil, response.InternalServerError("Failed to capture hold", err)
	}

	// Invalidate cache
	s.invalidateWalletCache(ctx, userID)

	logger.Info("hold captured", "userID", userID, "holdID", req.HoldID, "amount", captureAmount)

	return dto.ToTransactionResponse(transaction), nil
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
func (s *service) GetTransaction(ctx context.Context, userID, txID string) (*dto.TransactionResponse, error) {
	transaction, err := s.repo.FindTransactionByID(ctx, txID)
	if err != nil {
		return nil, response.NotFoundError("Transaction")
	}

	// Verify ownership
	wallet, err := s.repo.FindWalletByID(ctx, transaction.WalletID)
	if err != nil || wallet.UserID != userID {
		return nil, response.ForbiddenError("Not authorized to view this transaction")
	}

	return dto.ToTransactionResponse(transaction), nil
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

// CreditWallet - Internal method for other modules
func (s *service) CreditWallet(ctx context.Context, userID string, amount float64, refType, refID, description string, metadata map[string]interface{}) (*models.WalletTransaction, error) {
	walletResp, err := s.GetWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	wallet, err := s.repo.FindWalletByID(ctx, walletResp.ID)
	if err != nil {
		return nil, err
	}

	var transaction *models.WalletTransaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		balanceBefore := wallet.Balance
		wallet.Balance += amount

		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		now := time.Now()
		transaction = &models.WalletTransaction{
			WalletID:      wallet.ID,
			Type:          models.TransactionTypeCredit,
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

// Helper functions
func (s *service) invalidateWalletCache(ctx context.Context, userID string) {
	cache.Delete(ctx, fmt.Sprintf("wallet:user:%s", userID))
}

func stringPtr(s string) *string {
	return &s
}
