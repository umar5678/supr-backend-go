package customer

import (
	"context"
)

// WalletService defines the interface for wallet operations
// This is a placeholder interface - implement based on your actual wallet module
type WalletService interface {
	// GetBalance returns the current wallet balance for a user
	GetBalance(ctx context.Context, userID string) (float64, error)

	// HoldFunds places a hold on funds for a pending transaction
	// Returns hold ID if successful
	HoldFunds(ctx context.Context, userID string, amount float64, referenceType, referenceID, description string) (string, error)

	// ReleaseHold releases a held amount back to available balance
	ReleaseHold(ctx context.Context, holdID string) error

	// CaptureHold captures a held amount (converts hold to actual debit)
	CaptureHold(ctx context.Context, holdID string, amount float64, description string) error

	// Debit directly debits amount from wallet (for cancellation fees, etc.)
	Debit(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string) error

	// Credit credits amount to wallet (for refunds)
	Credit(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string) error
}

// MockWalletService is a placeholder implementation for development
type MockWalletService struct{}

func NewMockWalletService() WalletService {
	return &MockWalletService{}
}

func (m *MockWalletService) GetBalance(ctx context.Context, userID string) (float64, error) {
	// Return a mock balance for development
	return 10000.00, nil
}

func (m *MockWalletService) HoldFunds(ctx context.Context, userID string, amount float64, referenceType, referenceID, description string) (string, error) {
	// Return a mock hold ID
	return "hold_" + referenceID, nil
}

func (m *MockWalletService) ReleaseHold(ctx context.Context, holdID string) error {
	return nil
}

func (m *MockWalletService) CaptureHold(ctx context.Context, holdID string, amount float64, description string) error {
	return nil
}

func (m *MockWalletService) Debit(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string) error {
	return nil
}

func (m *MockWalletService) Credit(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string) error {
	return nil
}
 