package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/artyom/wallet-test-task/internal/models"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, models.ErrInsufficientFunds),
		errors.Is(err, models.ErrInvalidAmount),
		errors.Is(err, models.ErrInvalidOperation),
		errors.Is(err, models.ErrInvalidWalletID):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
