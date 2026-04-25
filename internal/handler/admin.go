package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Wild-sergunys/flowmodel/internal/model"
	"github.com/Wild-sergunys/flowmodel/internal/repository"
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

func (h *AdminHandler) GetMaterialByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	m, err := h.materialRepo.FindByID(r.Context(), id)
	switch {
	case err != nil:
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	case m == nil:
		WriteError(w, http.StatusNotFound, "not_found", "Материал не найден", nil)
		return
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

func (h *AdminHandler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	var m model.Material
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	if m.Name == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Название обязательно", nil)
		return
	}

	if err := h.materialRepo.Create(r.Context(), &m); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка создания", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": m.ID})
}

func (h *AdminHandler) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	var m model.Material
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}
	m.ID = id

	if err := h.materialRepo.Update(r.Context(), &m); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка обновления", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Материал обновлён"})
}

func (h *AdminHandler) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	if err := h.materialRepo.Delete(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка удаления", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Материал удалён"})
}
