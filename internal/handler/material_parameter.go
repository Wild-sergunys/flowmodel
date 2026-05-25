package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Wild-sergunys/flowmodel/internal/repository"
)

type MaterialParameterHandler struct {
	repo repository.MaterialParameterRepository
}

func NewMaterialParameterHandler(repo repository.MaterialParameterRepository) *MaterialParameterHandler {
	return &MaterialParameterHandler{repo: repo}
}

func (h *MaterialParameterHandler) ListParameters(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID материала", nil)
		return
	}

	params, err := h.repo.FindDetailsByMaterialID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка получения параметров", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(params)
}

func (h *MaterialParameterHandler) UpdateParameters(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	materialID, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID материала", nil)
		return
	}

	var req struct {
		Parameters []struct {
			Code       string  `json:"code"`
			ValueFloat float64 `json:"value_float"`
		} `json:"parameters"`
		Values map[string]float64 `json:"values"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	params := make(map[string]float64)

	for _, p := range req.Parameters {
		if p.Code != "" {
			params[p.Code] = p.ValueFloat
		}
	}

	for code, value := range req.Values {
		params[code] = value
	}

	if err := h.repo.Update(r.Context(), materialID, params); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сохранения параметров", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Параметры обновлены"})
}
