package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Wild-sergunys/flowmodel/internal/model"
	"github.com/Wild-sergunys/flowmodel/internal/repository"
)

type ParameterHandler struct {
	repo repository.ParameterRepository
}

func NewParameterHandler(repo repository.ParameterRepository) *ParameterHandler {
	return &ParameterHandler{repo: repo}
}

func (h *ParameterHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	params, err := h.repo.FindAll(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(params)
}

func (h *ParameterHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	p, err := h.repo.FindByID(r.Context(), id)
	switch {
	case err != nil:
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	case p == nil:
		WriteError(w, http.StatusNotFound, "not_found", "Параметр не найден", nil)
		return
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}
}

func (h *ParameterHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p model.Parameter
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	if err := h.repo.Create(r.Context(), &p); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка создания", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": p.ID})
}

func (h *ParameterHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	var p model.Parameter
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}
	p.ID = id

	if err := h.repo.Update(r.Context(), &p); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка обновления", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Параметр обновлён"})
}

func (h *ParameterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка удаления", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Параметр удалён"})
}
