package validator

import (
	"context"

	"flowmodel/internal/model"
	"flowmodel/internal/repository"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ValidateCalculationInput(input *model.CalculationInput, materialRepo repository.MaterialRepository) []ValidationError {
	var errors []ValidationError

	// Полнота
	if input.W == 0 {
		errors = append(errors, ValidationError{Field: "w", Message: "обязательное поле"})
	}
	if input.H == 0 {
		errors = append(errors, ValidationError{Field: "h", Message: "обязательное поле"})
	}
	if input.L == 0 {
		errors = append(errors, ValidationError{Field: "l", Message: "обязательное поле"})
	}
	if input.MaterialID == 0 {
		errors = append(errors, ValidationError{Field: "material_id", Message: "обязательное поле"})
	}
	if input.Steps == 0 {
		errors = append(errors, ValidationError{Field: "steps", Message: "обязательное поле"})
	}

	// Проверка существования материала
	if input.MaterialID != 0 {
		material, _ := materialRepo.FindByID(context.Background(), input.MaterialID)
		if material == nil {
			errors = append(errors, ValidationError{Field: "material_id", Message: "материал не найден"})
		}
	}

	// Диапазоны
	if input.W != 0 && (input.W < 0.001 || input.W > 10.0) {
		errors = append(errors, ValidationError{Field: "w", Message: "должно быть от 0.001 до 10.0"})
	}
	if input.H != 0 && (input.H < 0.001 || input.H > 1.0) {
		errors = append(errors, ValidationError{Field: "h", Message: "должно быть от 0.001 до 1.0"})
	}
	if input.L != 0 && (input.L < 0.01 || input.L > 100.0) {
		errors = append(errors, ValidationError{Field: "l", Message: "должно быть от 0.01 до 100.0"})
	}
	if input.Vu < 0 || input.Vu > 100.0 {
		errors = append(errors, ValidationError{Field: "vu", Message: "должно быть от 0 до 100.0"})
	}
	if input.Tu < 0 || input.Tu > 500.0 {
		errors = append(errors, ValidationError{Field: "tu", Message: "должно быть от 0 до 500.0"})
	}
	if input.T0 < 0 || input.T0 > 500.0 {
		errors = append(errors, ValidationError{Field: "t0", Message: "должно быть от 0 до 500.0"})
	}
	if input.Steps != 0 && (input.Steps < 10 || input.Steps > 100000) {
		errors = append(errors, ValidationError{Field: "steps", Message: "должно быть от 10 до 100000"})
	}

	// Логическая совместимость
	if input.T0 != 0 && input.Tu != 0 && input.T0 >= input.Tu {
		errors = append(errors, ValidationError{Field: "t0", Message: "начальная температура должна быть меньше температуры крышки"})
	}

	return errors
}
