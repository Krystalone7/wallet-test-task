package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestOperationRequest_WalletId(t *testing.T) {
	id := uuid.New()
	raw := []byte(`{"walletId":"` + id.String() + `","operationType":"DEPOSIT","amount":10}`)

	var req OperationRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatal(err)
	}
	if req.WalletID != id {
		t.Fatalf("got %s", req.WalletID)
	}
	if req.OperationType != OperationDeposit {
		t.Fatalf("got %s", req.OperationType)
	}
}
