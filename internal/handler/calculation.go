package handler

import (
	"encoding/json"
	"net/http"

	"flowmodel/internal/model"
	"flowmodel/internal/validator"
)

type CalculationHandler struct{}

func NewCalculationHandler() *CalculationHandler {
	return &CalculationHandler{}
}

func (h *CalculationHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var input model.CalculationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	errors := validator.ValidateCalculationInput(&input)
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
