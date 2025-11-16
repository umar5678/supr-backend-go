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
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	ReferenceType string  `json:"referenceType" binding:"required"`
	ReferenceID   string  `json:"referenceId" binding:"required,uuid"`
	HoldDuration  int     `json:"holdDuration" binding:"omitempty,min=1"` // minutes, default 30
}

func (r *HoldFundsRequest) Validate() error {
	if r.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if r.ReferenceType == "" {
		return errors.New("referenceType is required")
	}
	if r.ReferenceID == "" {
		return errors.New("referenceId is required")
	}
	if r.HoldDuration == 0 {
		r.HoldDuration = 30 // default 30 minutes
	}
	return nil
}

type ReleaseHoldRequest struct {
	HoldID string `json:"holdId" binding:"required,uuid"`
}

type CaptureHoldRequest struct {
	HoldID      string   `json:"holdId" binding:"required,uuid"`
	Amount      *float64 `json:"amount" binding:"omitempty,gt=0"` // Optional: capture partial amount
	Description string   `json:"description" binding:"omitempty"`
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
