.PHONY: up down test test-integration

up:
	docker compose up --build -d

down:
	docker compose down -v

test:
	go test ./...

test-integration:
	INTEGRATION=1 TEST_DB_HOST=127.0.0.1 TEST_DB_PORT=5433 go test ./internal/repository -count=1 -v
