package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Wild-sergunys/flowmodel/internal/middleware"
	"github.com/Wild-sergunys/flowmodel/internal/model"
	"github.com/Wild-sergunys/flowmodel/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

type CalculationHandler struct {
	calcService *service.CalculationService
}

func NewCalculationHandler(calcService *service.CalculationService) *CalculationHandler {
	return &CalculationHandler{
		calcService: calcService,
	}
}

func (h *CalculationHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var input model.CalculationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	errors := h.calcService.Validate(r.Context(), &input)
	if len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "validation_error",
			"message": "Ошибка валидации",
			"details": map[string]interface{}{"fields": errors},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Валидация пройдена"})
}

func (h *CalculationHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var input model.CalculationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}
	userID := int(claims["user_id"].(float64))

	result, err := h.calcService.Calculate(r.Context(), &input, userID)
	if err != nil {
		if verr, ok := err.(*service.ValidationError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "validation_error",
				"message": "Ошибка валидации",
				"details": map[string]interface{}{"fields": verr.Errors},
			})
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка расчёта", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *CalculationHandler) CalculateSurface(w http.ResponseWriter, r *http.Request) {
	var input model.CalculationSurfaceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	_, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

	result, err := h.calcService.CalculateSurface(r.Context(), &input)
	if err != nil {
		if verr, ok := err.(*service.ValidationError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "validation_error",
				"message": "Ошибка валидации",
				"details": map[string]interface{}{"fields": verr.Errors},
			})
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка расчёта поверхности", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
