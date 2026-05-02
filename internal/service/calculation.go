package service

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/Wild-sergunys/flowmodel/internal/model"
	"github.com/Wild-sergunys/flowmodel/internal/repository"
	"github.com/Wild-sergunys/flowmodel/internal/validator"
)

type CalculationService struct {
	materialParamRepo repository.MaterialParameterRepository
	materialRepo      repository.MaterialRepository
	calcRepo          repository.CalculationRepository
}

func NewCalculationService(
	materialParamRepo repository.MaterialParameterRepository,
	materialRepo repository.MaterialRepository,
	calcRepo repository.CalculationRepository,
) *CalculationService {
	return &CalculationService{
		materialParamRepo: materialParamRepo,
		materialRepo:      materialRepo,
		calcRepo:          calcRepo,
	}
}

func (s *CalculationService) Validate(ctx context.Context, input *model.CalculationInput) []validator.ValidationError {
	return validator.ValidateCalculationInput(input, s.materialRepo)
}

func (s *CalculationService) Calculate(ctx context.Context, input *model.CalculationInput, userID int) (*model.CalculationResult, error) {
	var result *model.CalculationResult
	var err error

	// До 3 попыток с экспоненциальной задержкой
	for attempt := 0; attempt < 3; attempt++ {
		result, err = s.calculateOnce(ctx, input, userID)
		if err == nil {
			return result, nil
		}

		// Если ошибка валидации - нет смысла повторять
		if _, ok := err.(*ValidationError); ok {
			return nil, err
		}

		// Если контекст отменён - выходим сразу
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Задержка перед повтором: 5ms, 25ms, 125ms
		if attempt < 2 {
			time.Sleep(time.Millisecond * time.Duration(5*(attempt+1)*(attempt+1)))
		}
	}

	return nil, err
}

func (s *CalculationService) calculateOnce(ctx context.Context, input *model.CalculationInput, userID int) (*model.CalculationResult, error) {
	errors := s.Validate(ctx, input)
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	params, err := s.materialParamRepo.FindByMaterialID(ctx, input.MaterialID)
	if err != nil {
		return nil, err
	}

	result := s.compute(input, params)

	inputJSON, _ := json.Marshal(input)
	resultJSON, _ := json.Marshal(result)

	calc := &model.Calculation{
		UserID:     userID,
		MaterialID: input.MaterialID,
		InputJSON:  string(inputJSON),
		ResultJSON: string(resultJSON),
	}

	if err := s.calcRepo.Create(ctx, calc); err != nil {
		return nil, err
	}

	result.ID = calc.ID
	return result, nil
}

func (s *CalculationService) compute(input *model.CalculationInput, params map[string]float64) *model.CalculationResult {
	mu0 := params["mu0"]
	Ea := params["Ea"]
	if Ea <= 0 {
		panic("отсутствует или некорректно задано значение энергии активации Ea")
	}
	Tr := params["Tr"]
	n := params["n"]
	alphaU := params["alpha_u"]
	density := params["density"]
	heatCapacity := params["heat_capacity"]
	meltingTemp := params["melting_temp"]

	if mu0 <= 0 {
		mu0 = 12000
	}
	if n <= 0 || n >= 1 {
		n = 0.28
	}
	if meltingTemp <= 0 {
		meltingTemp = 145
	}
	if Tr <= 0 {
		Tr = 180
	}
	if density <= 0 {
		density = 1380
	}
	if heatCapacity <= 0 {
		heatCapacity = 2500
	}
	if alphaU < 0 {
		alphaU = 400
	}

	const R = 8.314

	W := input.W
	H := input.H
	L := input.L
	Vu := input.Vu
	Tu := input.Tu

	if H <= 0 {
		panic("глубина канала H не может быть меньше или равна нулю")
	}
	if input.Steps <= 0 {
		input.Steps = 100 // fallback на значение по умолчанию
	}

	var F float64
	if W > 0 && H > 0 {
		aspectRatio := H / W
		if aspectRatio > 1 {
			aspectRatio = W / H
		}
		F = 1.0 - 0.628*aspectRatio
		if F < 0.3 {
			F = 0.3
		}
	} else {
		F = 1.0
	}

	gammaDot := Vu / H

	dz := L / float64(input.Steps)

	var profile []model.Point
	currentTemp := meltingTemp // Начальное условие: T(0) = T_0

	// Знаменатель для уравнения теплового баланса (rho * c * V * H)
	denominator := density * heatCapacity * Vu * H
	if math.Abs(denominator) < 1e-10 {
		denominator = 1
	}

	for i := 0; i <= input.Steps; i++ {
		z := float64(i) * dz

		// 1. Расчет температур в Кельвинах для реологического уравнения
		Tk := currentTemp + 273.15
		Trk := Tr + 273.15

		// 2. Расчет коэффициента консистенции по уравнению Андраде
		var mu float64
		if Ea > 0 && Tr > 0 {
			mu = mu0 * math.Exp(Ea/R*(1.0/Tk-1.0/Trk))
		} else {
			mu = mu0
		}

		// 3. Вычисление эффективной вязкости в текущем сечении
		viscosity := mu * math.Pow(gammaDot, n-1)

		// 4. Сохранение параметров текущего состояния
		profile = append(profile, model.Point{
			X:           z,
			Temperature: currentTemp,
			Viscosity:   viscosity,
		})

		// 5. Расчет локальных тепловых потоков (на основе текущей температуры currentTemp)
		currentQGamma := mu * math.Pow(gammaDot, n+1)
		currentQAlpha := alphaU * (Tu - currentTemp)
		currentQSum := currentQGamma + currentQAlpha

		// 6. Численное интегрирование: расчет приращения температуры dT на шаге dz
		dT := (currentQSum / denominator) * dz
		currentTemp += dT
	}

	lastPoint := profile[len(profile)-1]

	Qch := (W * H * Vu) / 2.0 * F
	Q := density * Qch * 3600

	// TODO: надо норм посчитать, по настоящему
	calcTimeMs := input.Steps * 2
	memoryBytes := input.Steps * 4096

	return &model.CalculationResult{
		Productivity: Q,
		Temperature:  lastPoint.Temperature,
		Viscosity:    lastPoint.Viscosity,
		Profile:      profile,
		Metrics: model.Metrics{
			CalcTimeMs:      calcTimeMs,
			MemoryUsedBytes: memoryBytes,
		},
	}
}

type ValidationError struct {
	Errors []validator.ValidationError
}

func (e *ValidationError) Error() string {
	return "ошибка валидации"
}
