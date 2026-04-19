package handler

import (
	"encoding/json"
	"net/http"

	"flowmodel/internal/repository"
)

type AdminHandler struct {
	materialRepo repository.MaterialRepository
}

func NewAdminHandler(materialRepo repository.MaterialRepository) *AdminHandler {
	return &AdminHandler{materialRepo: materialRepo}
}

func (h *AdminHandler) GetAllMaterials(w http.ResponseWriter, r *http.Request) {
	materials, err := h.materialRepo.FindAll(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}
