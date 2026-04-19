package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"flowmodel/internal/repository"
)

type ResultsHandler struct {
	calcRepo repository.CalculationRepository
}

func NewResultsHandler(calcRepo repository.CalculationRepository) *ResultsHandler {
	return &ResultsHandler{calcRepo: calcRepo}
}

func (h *ResultsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	calcs, err := h.calcRepo.FindAll(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calcs)
}

func (h *ResultsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	calc, err := h.calcRepo.FindByID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}
	if calc == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Расчёт не найден", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calc)
}

func (h *ResultsHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	calc, err := h.calcRepo.FindByID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}
	if calc == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Расчёт не найден", nil)
		return
	}

	report := map[string]interface{}{
		"id":          calc.ID,
		"user_id":     calc.UserID,
		"material_id": calc.MaterialID,
		"created_at":  calc.CreatedAt,
		"input":       json.RawMessage(calc.InputJSON),
		"result":      json.RawMessage(calc.ResultJSON),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ResultsHandler) Download(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	calc, err := h.calcRepo.FindByID(r.Context(), id)
	if err != nil || calc == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Расчёт не найден", nil)
		return
	}

	report := map[string]interface{}{
		"id":          calc.ID,
		"user_id":     calc.UserID,
		"material_id": calc.MaterialID,
		"created_at":  calc.CreatedAt,
		"input":       json.RawMessage(calc.InputJSON),
		"result":      json.RawMessage(calc.ResultJSON),
	}

	data, _ := json.MarshalIndent(report, "", "  ")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=report_"+strconv.Itoa(id)+".json")
	w.Write(data)
}
