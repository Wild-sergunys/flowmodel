package handler

import (
	"encoding/json"
	"github.com/Wild-sergunys/flowmodel/internal/model"
	"net/http"
)

func WriteError(w http.ResponseWriter, status int, errorType, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.ErrorResponse{
		Error:   errorType,
		Message: message,
		Details: details,
	})
}
