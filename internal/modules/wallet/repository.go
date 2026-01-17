package wallet

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	// Wallet operations
	CreateWallet(ctx context.Context, wallet *models.Wallet) error
	FindWalletByID(ctx context.Context, id string) (*models.Wallet, error)
	FindWalletByUserID(ctx context.Context, userID string, walletType models.WalletType) (*models.Wallet, error)
	UpdateWallet(ctx context.Context, wallet *models.Wallet) error

	// Transaction operations
	CreateTransaction(ctx context.Context, tx *models.WalletTransaction) error
	FindTransactionByID(ctx context.Context, id string) (*models.WalletTransaction, error)
	ListTransactions(ctx context.Context, walletID string, filters map[string]interface{}, page, limit int) ([]*models.WalletTransaction, int64, error)

	// Hold operations
	CreateHold(ctx context.Context, hold *models.WalletHold) error
	FindHoldByID(ctx context.Context, id string) (*models.WalletHold, error)
	FindHoldsByReference(ctx context.Context, refType, refID string) ([]*models.WalletHold, error)
	UpdateHold(ctx context.Context, hold *models.WalletHold) error
	ReleaseExpiredHolds(ctx context.Context) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *repository) FindWalletByID(ctx context.Context, id string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&wallet).Error
	return &wallet, err
}

func (r *repository) FindWalletByUserID(ctx context.Context, userID string, walletType models.WalletType) (*models.Wallet, error) {
	var wallet models.Wallet
	// First try: find wallet with matching type
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ? AND wallet_type = ?", userID, walletType).
		First(&wallet).Error

	// Fallback: if not found and no records, try finding ANY wallet for user (for legacy data)
	if gorm.ErrRecordNotFound == err {
		err = r.db.WithContext(ctx).
			Preload("User").
			Where("user_id = ?", userID).
			First(&wallet).Error

		// If found via fallback, update it to the correct wallet type
		if err == nil {
			wallet.WalletType = walletType
			r.UpdateWallet(ctx, &wallet)
		}
	}

	return &wallet, err
}

func (r *repository) UpdateWallet(ctx context.Context, wallet *models.Wallet) error {
	return r.db.WithContext(ctx).Save(wallet).Error
}

func (r *repository) CreateTransaction(ctx context.Context, tx *models.WalletTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *repository) FindTransactionByID(ctx context.Context, id string) (*models.WalletTransaction, error) {
	var tx models.WalletTransaction
	err := r.db.WithContext(ctx).
		Preload("Wallet").
		Where("id = ?", id).
		First(&tx).Error
	return &tx, err
}

func (r *repository) ListTransactions(ctx context.Context, walletID string, filters map[string]interface{}, page, limit int) ([]*models.WalletTransaction, int64, error) {
	var transactions []*models.WalletTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.WalletTransaction{}).
		Where("wallet_id = ?", walletID)

	// Apply filters
	if txType, ok := filters["type"].(models.TransactionType); ok && txType != "" {
		query = query.Where("transaction_type = ?", txType)
	}
	if status, ok := filters["status"].(models.TransactionStatus); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	query.Count(&total)

	// Paginate
	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions).Error

	return transactions, total, err
}

func (r *repository) CreateHold(ctx context.Context, hold *models.WalletHold) error {
	return r.db.WithContext(ctx).Create(hold).Error
}

func (r *repository) FindHoldByID(ctx context.Context, id string) (*models.WalletHold, error) {
	var hold models.WalletHold
	err := r.db.WithContext(ctx).
		Preload("Wallet").
		Where("id = ?", id).
		First(&hold).Error
	return &hold, err
}

func (r *repository) FindHoldsByReference(ctx context.Context, refType, refID string) ([]*models.WalletHold, error) {
	var holds []*models.WalletHold
	err := r.db.WithContext(ctx).
		Where("reference_type = ? AND reference_id = ? AND status = ?", refType, refID, models.TransactionStatusHeld).
		Find(&holds).Error
	return holds, err
}

func (r *repository) UpdateHold(ctx context.Context, hold *models.WalletHold) error {
	return r.db.WithContext(ctx).Save(hold).Error
}

func (r *repository) ReleaseExpiredHolds(ctx context.Context) error {
	now := time.Now()
	var expiredHolds []*models.WalletHold

	// Find expired holds
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at < ?", models.TransactionStatusHeld, now).
		Find(&expiredHolds).Error
	if err != nil {
		return err
	}

	// Release each hold
	for _, hold := range expiredHolds {
		// Update wallet
		var wallet models.Wallet
		if err := r.db.WithContext(ctx).Where("id = ?", hold.WalletID).First(&wallet).Error; err != nil {
			continue
		}

		wallet.HeldBalance -= hold.Amount

		// Update in transaction
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&wallet).Error; err != nil {
				return err
			}

			now := time.Now()
			hold.Status = models.TransactionStatusReleased
			hold.ReleasedAt = &now
			if err := tx.Save(hold).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			continue
		}
	}

	return nil
}
