package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/artyom/wallet-test-task/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) GetBalance(ctx context.Context, id uuid.UUID) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx, `SELECT balance FROM wallets WHERE id = $1`, id).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, models.ErrWalletNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	return balance, nil
}

func (r *WalletRepository) Deposit(ctx context.Context, id uuid.UUID, amount int64) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO wallets (id, balance)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE
		SET balance = wallets.balance + EXCLUDED.balance,
		    updated_at = NOW()
		RETURNING balance
	`, id, amount).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("deposit: %w", err)
	}
	return balance, nil
}

func (r *WalletRepository) Withdraw(ctx context.Context, id uuid.UUID, amount int64) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx, `
		UPDATE wallets
		SET balance = balance - $2,
		    updated_at = NOW()
		WHERE id = $1 AND balance >= $2
		RETURNING balance
	`, id, amount).Scan(&balance)

	if errors.Is(err, pgx.ErrNoRows) {
		exists, checkErr := r.exists(ctx, id)
		if checkErr != nil {
			return 0, checkErr
		}
		if !exists {
			return 0, models.ErrWalletNotFound
		}
		return 0, models.ErrInsufficientFunds
	}
	if err != nil {
		return 0, fmt.Errorf("withdraw: %w", err)
	}
	return balance, nil
}

func (r *WalletRepository) exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM wallets WHERE id = $1)`, id).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("check wallet: %w", err)
	}
	return ok, nil
}
