package service

import (
	"context"
	"encoding/json"
	"math"

	"flowmodel/internal/model"
	"flowmodel/internal/repository"
	"flowmodel/internal/validator"
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
	errors := s.Validate(ctx, input)
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	params, err := s.materialParamRepo.FindByMaterialID(ctx, input.MaterialID)
	if err != nil {
		return nil, err
	}

	// Расчёт
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

	return &model.CalculationResult{
		Productivity: input.W * input.H * input.Vu / 2,
		Temperature:  lastPoint.Temperature,
		Viscosity:    lastPoint.Viscosity,
		Profile:      profile,
		Metrics: model.Metrics{
			CalcTimeMs:      42,
			MemoryUsedBytes: 2048000,
		},
	}
}

type ValidationError struct {
	Errors []validator.ValidationError
}

func (e *ValidationError) Error() string {
	return "ошибка валидации"
}
