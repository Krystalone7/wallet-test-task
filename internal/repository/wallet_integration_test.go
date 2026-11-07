package repository

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/artyom/wallet-test-task/internal/config"
	"github.com/artyom/wallet-test-task/internal/models"
	"github.com/google/uuid"
)

func testRepo(t *testing.T) *WalletRepository {
	t.Helper()

	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("set INTEGRATION=1 to run db tests")
	}

	cfg, err := config.Load("../../config.env")
	if err != nil {
		t.Fatal(err)
	}
	cfg.DBHost = "127.0.0.1"
	cfg.DBPort = "5433"
	if h := os.Getenv("TEST_DB_HOST"); h != "" {
		cfg.DBHost = h
	}
	if p := os.Getenv("TEST_DB_PORT"); p != "" {
		cfg.DBPort = p
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("db unavailable: %v", err)
	}
	t.Cleanup(pool.Close)

	return NewWalletRepository(pool)
}

func TestDepositAndGetBalance(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	id := uuid.New()

	bal, err := repo.Deposit(ctx, id, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if bal != 1000 {
		t.Fatalf("got %d", bal)
	}

	got, err := repo.GetBalance(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1000 {
		t.Fatalf("got %d", got)
	}
}

func TestWithdrawInsufficient(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	id := uuid.New()

	if _, err := repo.Deposit(ctx, id, 100); err != nil {
		t.Fatal(err)
	}

	_, err := repo.Withdraw(ctx, id, 500)
	if err != models.ErrInsufficientFunds {
		t.Fatalf("expected insufficient funds, got %v", err)
	}
}

func TestConcurrentWithdrawals(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	id := uuid.New()

	const startBalance int64 = 1000
	if _, err := repo.Deposit(ctx, id, startBalance); err != nil {
		t.Fatal(err)
	}

	var (
		wg       sync.WaitGroup
		okCount  atomic.Int64
		failSoft atomic.Int64
		hardErr  atomic.Int64
	)

	workers := 2000
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, err := repo.Withdraw(ctx, id, 1)
			switch {
			case err == nil:
				okCount.Add(1)
			case err == models.ErrInsufficientFunds:
				failSoft.Add(1)
			default:
				hardErr.Add(1)
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if hardErr.Load() != 0 {
		t.Fatalf("got hard errors: %d", hardErr.Load())
	}
	if okCount.Load() != startBalance {
		t.Fatalf("successful withdraws=%d, want %d (soft fails=%d)", okCount.Load(), startBalance, failSoft.Load())
	}
	if failSoft.Load() != int64(workers)-startBalance {
		t.Fatalf("soft fails=%d, want %d", failSoft.Load(), int64(workers)-startBalance)
	}

	bal, err := repo.GetBalance(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if bal != 0 {
		t.Fatalf("final balance=%d, want 0", bal)
	}
}

func TestConcurrentDeposits(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	id := uuid.New()

	workers := 500
	var wg sync.WaitGroup
	wg.Add(workers)

	var hardErr atomic.Int64
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if _, err := repo.Deposit(ctx, id, 1); err != nil {
				hardErr.Add(1)
				t.Errorf("deposit failed: %v", err)
			}
		}()
	}
	wg.Wait()

	if hardErr.Load() != 0 {
		t.Fatalf("hard errors: %d", hardErr.Load())
	}

	bal, err := repo.GetBalance(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if bal != int64(workers) {
		t.Fatalf("balance=%d, want %d", bal, workers)
	}
}
