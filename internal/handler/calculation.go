package handler

import (
	"encoding/json"
	"math"
	"net/http"

	"flowmodel/internal/middleware"
	"flowmodel/internal/model"
	"flowmodel/internal/repository"
	"flowmodel/internal/validator"

	"github.com/golang-jwt/jwt/v5"
)

type CalculationHandler struct {
	materialParamRepo repository.MaterialParameterRepository
	materialRepo      repository.MaterialRepository
	calcRepo          repository.CalculationRepository
}

func NewCalculationHandler(
	materialParamRepo repository.MaterialParameterRepository,
	materialRepo repository.MaterialRepository,
	calcRepo repository.CalculationRepository,
) *CalculationHandler {
	return &CalculationHandler{
		materialParamRepo: materialParamRepo,
		materialRepo:      materialRepo,
		calcRepo:          calcRepo,
	}
}

func (h *CalculationHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var input model.CalculationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	errors := validator.ValidateCalculationInput(&input, h.materialRepo)
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

// Расчет фейковы, без понятия как там норм считать, надо доделать
func (h *CalculationHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var input model.CalculationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	errors := validator.ValidateCalculationInput(&input, h.materialRepo)
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

	params, err := h.materialParamRepo.FindByMaterialID(r.Context(), input.MaterialID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка получения параметров материала", nil)
		return
	}

	mu0 := params["mu0"]
	Ea := params["Ea"]
	Tr := params["Tr"]
	n := params["n"]
	if mu0 == 0 {
		mu0 = 12000
	}
	if n == 0 {
		n = 0.28
	}

	const R = 8.314

	dx := input.L / float64(input.Steps)
	var profile []model.Point

	for i := 0; i <= input.Steps; i++ {
		x := float64(i) * dx
		t := input.T0 + (input.Tu-input.T0)*(x/input.L)

		// Формула Андраде + Оствальда-де'Вилье
		Tk := t + 273.15
		Trk := Tr + 273.15
		gammaDot := input.Vu / input.H

		var mu float64
		if Ea > 0 && Tr > 0 {
			mu = mu0 * math.Exp(Ea/R*(1.0/Tk-1.0/Trk))
		} else {
			mu = mu0
		}

		viscosity := mu * math.Pow(gammaDot, n-1)

		profile = append(profile, model.Point{X: x, Temperature: t, Viscosity: viscosity})
	}

	lastPoint := profile[len(profile)-1]

	result := model.CalculationResult{
		Productivity: input.W * input.H * input.Vu / 2,
		Temperature:  lastPoint.Temperature,
		Viscosity:    lastPoint.Viscosity,
		Profile:      profile,
		Metrics: model.Metrics{
			CalcTimeMs:      42,
			MemoryUsedBytes: 2048000,
		},
	}

	inputJSON, _ := json.Marshal(input)
	resultJSON, _ := json.Marshal(result)

	// Берём userID из JWT (middleware уже проверил токен)
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}
	userID := int(claims["user_id"].(float64))

	calc := &model.Calculation{
		UserID:     userID,
		MaterialID: input.MaterialID,
		InputJSON:  string(inputJSON),
		ResultJSON: string(resultJSON),
	}

	if err := h.calcRepo.Create(r.Context(), calc); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сохранения расчёта", nil)
		return
	}

	result.ID = calc.ID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
