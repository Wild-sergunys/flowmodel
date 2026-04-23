package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"

	"flowmodel/internal/middleware"
	"flowmodel/internal/model"
	"flowmodel/internal/repository"
)

type ResultsHandler struct {
	calcRepo repository.CalculationRepository
}

func NewResultsHandler(calcRepo repository.CalculationRepository) *ResultsHandler {
	return &ResultsHandler{calcRepo: calcRepo}
}

func getUserID(r *http.Request) (int, bool) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		return 0, false
	}
	userID := int(claims["user_id"].(float64))
	return userID, true
}

func (h *ResultsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

	calcs, err := h.calcRepo.FindByUserID(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}

	if calcs == nil {
		calcs = []model.Calculation{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calcs)
}

func (h *ResultsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

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

	if calc.UserID != userID {
		WriteError(w, http.StatusForbidden, "forbidden", "Нет доступа к этому расчёту", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calc)
}

func (h *ResultsHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

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

	if calc.UserID != userID {
		WriteError(w, http.StatusForbidden, "forbidden", "Нет доступа к этому расчёту", nil)
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
	userID, ok := getUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

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

	if calc.UserID != userID {
		WriteError(w, http.StatusForbidden, "forbidden", "Нет доступа к этому расчёту", nil)
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
