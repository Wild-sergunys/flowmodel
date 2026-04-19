package handler

import (
	"encoding/json"
	"net/http"

	"flowmodel/internal/repository"
)

type MaterialHandler struct {
	repo repository.MaterialRepository
}

func NewMaterialHandler(repo repository.MaterialRepository) *MaterialHandler {
	return &MaterialHandler{repo: repo}
}

func (h *MaterialHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	materials, err := h.repo.FindAll(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка при получении материалов", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}

