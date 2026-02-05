package dto

import (
	"errors"

	"github.com/umar5678/go-backend/internal/models"
)

type AddFundsRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"omitempty"`
}

func (r *AddFundsRequest) Validate() error {
	if r.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if r.Amount > 10000 {
		return errors.New("maximum amount is $10,000 per transaction")
	}
	return nil
}

type WithdrawFundsRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"omitempty"`
}

func (r *WithdrawFundsRequest) Validate() error {
	if r.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	return nil
}

type TransferFundsRequest struct {
	RecipientID string  `json:"recipientId" binding:"required,uuid"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"omitempty"`
}

func (r *TransferFundsRequest) Validate() error {
	if r.RecipientID == "" {
		return errors.New("recipientId is required")
	}
	if r.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	return nil
}

type HoldFundsRequest struct {
    Amount        float64 `json:"amount" binding:"required,min=0.5"`
    ReferenceType string  `json:"referenceType" binding:"required"`
    ReferenceID   string  `json:"referenceId" binding:"required"`
    HoldDuration  int     `json:"holdDuration" binding:"omitempty,min=60,max=3600"` // seconds
}

func (r *HoldFundsRequest) Validate() error {
    if r.HoldDuration == 0 {
        r.HoldDuration = 1800 // 30 minutes default
    }
    return nil
}

type ReleaseHoldRequest struct {
	HoldID string `json:"holdId" binding:"required,uuid"`
}

type CaptureHoldRequest struct {
    HoldID      string   `json:"holdId" binding:"required,uuid"`
    Amount      *float64 `json:"amount" binding:"omitempty,min=0.5"`
    Description string   `json:"description" binding:"omitempty,max=500"`
}

type TransactionHistoryRequest struct {
    Page      int    `form:"page" binding:"omitempty,min=1"`
    Limit     int    `form:"limit" binding:"omitempty,min=1,max=100"`
    Type      string `form:"type" binding:"omitempty,oneof=credit debit"`
    StartDate string `form:"startDate" binding:"omitempty"`
    EndDate   string `form:"endDate" binding:"omitempty"`
}

func (r *TransactionHistoryRequest) SetDefaults() {
    if r.Page == 0 {
        r.Page = 1
    }
    if r.Limit == 0 {
        r.Limit = 20
    }
}

type ListTransactionsRequest struct {
	Page   int                      `form:"page" binding:"omitempty,min=1"`
	Limit  int                      `form:"limit" binding:"omitempty,min=1,max=100"`
	Type   models.TransactionType   `form:"type" binding:"omitempty"`
	Status models.TransactionStatus `form:"status" binding:"omitempty"`
}

func (r *ListTransactionsRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 20
	}
}

type CashCollectionRequest struct {
    RideID string  `json:"rideId" binding:"required,uuid"`
    Amount float64 `json:"amount" binding:"required,min=0.5"`
}

func (r *CashCollectionRequest) Validate() error {
    if r.Amount < 0.5 {
        return errors.New("invalid amount")
    }
    return nil
}

type CashPaymentRequest struct {
    Amount       float64 `json:"amount" binding:"required,min=1"`
    SettlementID string  `json:"settlementId" binding:"required"`
}

func (r *CashPaymentRequest) Validate() error {
    if r.Amount < 1.0 {
        return errors.New("minimum settlement amount is $1.00")
    }
    return nil
}
