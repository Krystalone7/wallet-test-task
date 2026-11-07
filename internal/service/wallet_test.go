package service

import (
	"context"
	"testing"

	"github.com/artyom/wallet-test-task/internal/models"
	"github.com/google/uuid"
)

type mockRepo struct {
	balance int64
	exists  bool
}

func (m *mockRepo) GetBalance(ctx context.Context, id uuid.UUID) (int64, error) {
	if !m.exists {
		return 0, models.ErrWalletNotFound
	}
	return m.balance, nil
}

func (m *mockRepo) Deposit(ctx context.Context, id uuid.UUID, amount int64) (int64, error) {
	m.exists = true
	m.balance += amount
	return m.balance, nil
}

func (m *mockRepo) Withdraw(ctx context.Context, id uuid.UUID, amount int64) (int64, error) {
	if !m.exists {
		return 0, models.ErrWalletNotFound
	}
	if m.balance < amount {
		return 0, models.ErrInsufficientFunds
	}
	m.balance -= amount
	return m.balance, nil
}

func TestApplyOperation_Deposit(t *testing.T) {
	svc := NewWalletService(&mockRepo{})
	id := uuid.New()

	wallet, err := svc.ApplyOperation(context.Background(), models.OperationRequest{
		WalletID:      id,
		OperationType: models.OperationDeposit,
		Amount:        1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wallet.Balance != 1000 {
		t.Fatalf("expected balance 1000, got %d", wallet.Balance)
	}
}

func TestApplyOperation_WithdrawInsufficient(t *testing.T) {
	svc := NewWalletService(&mockRepo{exists: true, balance: 100})
	id := uuid.New()

	_, err := svc.ApplyOperation(context.Background(), models.OperationRequest{
		WalletID:      id,
		OperationType: models.OperationWithdraw,
		Amount:        500,
	})
	if err != models.ErrInsufficientFunds {
		t.Fatalf("expected insufficient funds, got %v", err)
	}
}

func TestApplyOperation_InvalidAmount(t *testing.T) {
	svc := NewWalletService(&mockRepo{})
	_, err := svc.ApplyOperation(context.Background(), models.OperationRequest{
		WalletID:      uuid.New(),
		OperationType: models.OperationDeposit,
		Amount:        0,
	})
	if err != models.ErrInvalidAmount {
		t.Fatalf("expected invalid amount, got %v", err)
	}
}

func TestApplyOperation_InvalidType(t *testing.T) {
	svc := NewWalletService(&mockRepo{})
	_, err := svc.ApplyOperation(context.Background(), models.OperationRequest{
		WalletID:      uuid.New(),
		OperationType: "TRANSFER",
		Amount:        10,
	})
	if err != models.ErrInvalidOperation {
		t.Fatalf("expected invalid operation, got %v", err)
	}
}

func TestGetBalance_NotFound(t *testing.T) {
	svc := NewWalletService(&mockRepo{})
	_, err := svc.GetBalance(context.Background(), uuid.New())
	if err != models.ErrWalletNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
