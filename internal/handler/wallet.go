package handler

import (
	"encoding/json"
	"net/http"

	"github.com/artyom/wallet-test-task/internal/models"
	"github.com/artyom/wallet-test-task/internal/service"
	"github.com/google/uuid"
)

type WalletHandler struct {
	svc *service.WalletService
}

func NewWalletHandler(svc *service.WalletService) *WalletHandler {
	return &WalletHandler{svc: svc}
}

func (h *WalletHandler) Operate(w http.ResponseWriter, r *http.Request) {
	var req models.OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	wallet, err := h.svc.ApplyOperation(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, wallet)
}

func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("walletId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: models.ErrInvalidWalletID.Error()})
		return
	}

	wallet, err := h.svc.GetBalance(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, wallet)
}

func (h *WalletHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
