package customer

import (
	"context"
)

type WalletService interface {
	GetBalance(ctx context.Context, userID string) (float64, error)
	HoldFunds(ctx context.Context, userID string, amount float64, referenceType, referenceID, description string) (string, error)
	ReleaseHold(ctx context.Context, holdID string) error
	CaptureHold(ctx context.Context, holdID string, amount float64, description string) error
	Debit(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string) error
	Credit(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string) error
}

type MockWalletService struct{}

func NewMockWalletService() WalletService {
	return &MockWalletService{}
}

func (m *MockWalletService) GetBalance(ctx context.Context, userID string) (float64, error) {
	return 10000.00, nil
}

func (m *MockWalletService) HoldFunds(ctx context.Context, userID string, amount float64, referenceType, referenceID, description string) (string, error) {
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
 