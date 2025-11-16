package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
)

type WalletResponse struct {
	ID               string                `json:"id"`
	UserID           string                `json:"userId"`
	WalletType       models.WalletType     `json:"walletType"`
	Balance          float64               `json:"balance"`
	HeldBalance      float64               `json:"heldBalance"`
	AvailableBalance float64               `json:"availableBalance"`
	Currency         string                `json:"currency"`
	IsActive         bool                  `json:"isActive"`
	CreatedAt        time.Time             `json:"createdAt"`
	UpdatedAt        time.Time             `json:"updatedAt"`
	User             *authdto.UserResponse `json:"user,omitempty"`
}

type TransactionResponse struct {
	ID            string                   `json:"id"`
	WalletID      string                   `json:"walletId"`
	Type          models.TransactionType   `json:"type"`
	Amount        float64                  `json:"amount"`
	BalanceBefore float64                  `json:"balanceBefore"`
	BalanceAfter  float64                  `json:"balanceAfter"`
	Status        models.TransactionStatus `json:"status"`
	ReferenceType *string                  `json:"referenceType,omitempty"`
	ReferenceID   *string                  `json:"referenceId,omitempty"`
	Description   *string                  `json:"description,omitempty"`
	Metadata      map[string]interface{}   `json:"metadata,omitempty"`
	ProcessedAt   *time.Time               `json:"processedAt,omitempty"`
	CreatedAt     time.Time                `json:"createdAt"`
}

type HoldResponse struct {
	ID            string                   `json:"id"`
	WalletID      string                   `json:"walletId"`
	Amount        float64                  `json:"amount"`
	ReferenceType string                   `json:"referenceType"`
	ReferenceID   string                   `json:"referenceId"`
	Status        models.TransactionStatus `json:"status"`
	ExpiresAt     time.Time                `json:"expiresAt"`
	ReleasedAt    *time.Time               `json:"releasedAt,omitempty"`
	CreatedAt     time.Time                `json:"createdAt"`
}

func ToWalletResponse(wallet *models.Wallet) *WalletResponse {
	resp := &WalletResponse{
		ID:               wallet.ID,
		UserID:           wallet.UserID,
		WalletType:       wallet.WalletType,
		Balance:          wallet.Balance,
		HeldBalance:      wallet.HeldBalance,
		AvailableBalance: wallet.GetAvailableBalance(),
		Currency:         wallet.Currency,
		IsActive:         wallet.IsActive,
		CreatedAt:        wallet.CreatedAt,
		UpdatedAt:        wallet.UpdatedAt,
	}

	if wallet.User.ID != "" {
		resp.User = authdto.ToUserResponse(&wallet.User)
	}

	return resp
}

func ToTransactionResponse(tx *models.WalletTransaction) *TransactionResponse {
	return &TransactionResponse{
		ID:            tx.ID,
		WalletID:      tx.WalletID,
		Type:          tx.Type,
		Amount:        tx.Amount,
		BalanceBefore: tx.BalanceBefore,
		BalanceAfter:  tx.BalanceAfter,
		Status:        tx.Status,
		ReferenceType: tx.ReferenceType,
		ReferenceID:   tx.ReferenceID,
		Description:   tx.Description,
		Metadata:      tx.Metadata,
		ProcessedAt:   tx.ProcessedAt,
		CreatedAt:     tx.CreatedAt,
	}
}

func ToHoldResponse(hold *models.WalletHold) *HoldResponse {
	return &HoldResponse{
		ID:            hold.ID,
		WalletID:      hold.WalletID,
		Amount:        hold.Amount,
		ReferenceType: hold.ReferenceType,
		ReferenceID:   hold.ReferenceID,
		Status:        hold.Status,
		ExpiresAt:     hold.ExpiresAt,
		ReleasedAt:    hold.ReleasedAt,
		CreatedAt:     hold.CreatedAt,
	}
}
