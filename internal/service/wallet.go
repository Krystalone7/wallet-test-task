package service

import (
	"context"

	"github.com/artyom/wallet-test-task/internal/models"
	"github.com/google/uuid"
)

type WalletRepo interface {
	GetBalance(ctx context.Context, id uuid.UUID) (int64, error)
	Deposit(ctx context.Context, id uuid.UUID, amount int64) (int64, error)
	Withdraw(ctx context.Context, id uuid.UUID, amount int64) (int64, error)
}

type WalletService struct {
	repo WalletRepo
}

func NewWalletService(repo WalletRepo) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) ApplyOperation(ctx context.Context, req models.OperationRequest) (*models.Wallet, error) {
	if req.WalletID == uuid.Nil {
		return nil, models.ErrInvalidWalletID
	}
	if req.Amount <= 0 {
		return nil, models.ErrInvalidAmount
	}

	var (
		balance int64
		err     error
	)

	switch req.OperationType {
	case models.OperationDeposit:
		balance, err = s.repo.Deposit(ctx, req.WalletID, req.Amount)
	case models.OperationWithdraw:
		balance, err = s.repo.Withdraw(ctx, req.WalletID, req.Amount)
	default:
		return nil, models.ErrInvalidOperation
	}
	if err != nil {
		return nil, err
	}

	return &models.Wallet{ID: req.WalletID, Balance: balance}, nil
}

func (s *WalletService) GetBalance(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	if id == uuid.Nil {
		return nil, models.ErrInvalidWalletID
	}

	balance, err := s.repo.GetBalance(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.Wallet{ID: id, Balance: balance}, nil
}
