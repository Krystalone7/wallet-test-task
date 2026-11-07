package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artyom/wallet-test-task/internal/models"
	"github.com/artyom/wallet-test-task/internal/service"
	"github.com/google/uuid"
)

type stubRepo struct {
	balance map[uuid.UUID]int64
}

func newStubRepo() *stubRepo {
	return &stubRepo{balance: make(map[uuid.UUID]int64)}
}

func (s *stubRepo) GetBalance(ctx context.Context, id uuid.UUID) (int64, error) {
	b, ok := s.balance[id]
	if !ok {
		return 0, models.ErrWalletNotFound
	}
	return b, nil
}

func (s *stubRepo) Deposit(ctx context.Context, id uuid.UUID, amount int64) (int64, error) {
	s.balance[id] += amount
	return s.balance[id], nil
}

func (s *stubRepo) Withdraw(ctx context.Context, id uuid.UUID, amount int64) (int64, error) {
	b, ok := s.balance[id]
	if !ok {
		return 0, models.ErrWalletNotFound
	}
	if b < amount {
		return 0, models.ErrInsufficientFunds
	}
	s.balance[id] = b - amount
	return s.balance[id], nil
}

func TestOperate_Deposit(t *testing.T) {
	h := NewWalletHandler(service.NewWalletService(newStubRepo()))
	id := uuid.New()

	body := map[string]any{
		"walletId":      id.String(),
		"operationType": "DEPOSIT",
		"amount":        1500,
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(raw))
	rr := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	var resp models.Wallet
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Balance != 1500 {
		t.Fatalf("balance=%d", resp.Balance)
	}
}

func TestGetBalance_NotFound(t *testing.T) {
	h := NewWalletHandler(service.NewWalletService(newStubRepo()))
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+id.String(), nil)
	rr := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestOperate_BusinessErrorIsNot500(t *testing.T) {
	repo := newStubRepo()
	id := uuid.New()
	repo.balance[id] = 50

	h := NewWalletHandler(service.NewWalletService(repo))
	body := map[string]any{
		"walletId":      id.String(),
		"operationType": "WITHDRAW",
		"amount":        100,
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(raw))
	rr := httptest.NewRecorder()
	NewRouter(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
