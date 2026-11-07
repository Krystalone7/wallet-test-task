# Wallet API

## API

### Изменить баланс

`POST /api/v1/wallet`

```json
{
  "walletId": "550e8400-e29b-41d4-a716-446655440000",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

`operationType`: `DEPOSIT` | `WITHDRAW`

### Получить баланс

`GET /api/v1/wallets/{WALLET_UUID}`

## Запуск

```bash
docker compose up --build
```

Сервис: `http://localhost:8080`  
Конфиг: `config.env`

## Примеры

```bash
# пополнить
curl -X POST http://localhost:8080/api/v1/wallet \
  -H "Content-Type: application/json" \
  -d "{\"walletId\":\"550e8400-e29b-41d4-a716-446655440000\",\"operationType\":\"DEPOSIT\",\"amount\":1000}"

# баланс
curl http://localhost:8080/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000
```

## Тесты

```bash
# юнит-тесты 
go test ./...

# интеграционные
docker compose up -d postgres
# powershell:
$env:INTEGRATION=1; go test ./internal/repository -count=1 -v
```

## Конкурентность

Операции делаются одним атомарным `UPDATE`/`INSERT ... ON CONFLICT`.
Postgres сам сериализует апдейты одной строки, баланс не уходит в минус (`CHECK` + условие `balance >= amount`).
