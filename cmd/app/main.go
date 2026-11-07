package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artyom/wallet-test-task/internal/config"
	"github.com/artyom/wallet-test-task/internal/handler"
	"github.com/artyom/wallet-test-task/internal/repository"
	"github.com/artyom/wallet-test-task/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load("config.env")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := waitForDB(ctx, cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	repo := repository.NewWalletRepository(pool)
	svc := service.NewWalletService(repo)
	h := handler.NewWalletHandler(svc)

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           handler.NewRouter(h),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func waitForDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	var lastErr error
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		p, err := repository.NewPool(ctx, cfg)
		if err == nil {
			return p, nil
		}
		lastErr = err
		log.Printf("db not ready yet: %v", err)

		select {
		case <-ctx.Done():
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
