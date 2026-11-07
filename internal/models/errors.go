package models

import "errors"

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrInvalidOperation    = errors.New("operationType must be DEPOSIT or WITHDRAW")
	ErrInvalidWalletID     = errors.New("invalid wallet id")
)
