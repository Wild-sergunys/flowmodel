package service

import (
	"context"
	"encoding/json"
	"math"
	"runtime"
	"sync"
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

func (s *CalculationService) ValidateSurface(input *model.CalculationSurfaceInput) []validator.ValidationError {
	var errors []validator.ValidationError

	if input.W <= 0 || input.W < 0.001 || input.W > 10.0 {
		errors = append(errors, validator.ValidationError{Field: "w", Message: "должно быть от 0.001 до 10.0"})
	}
	if input.H <= 0 || input.H < 0.001 || input.H > 1.0 {
		errors = append(errors, validator.ValidationError{Field: "h", Message: "должно быть от 0.001 до 1.0"})
	}
	if input.L <= 0 || input.L < 0.01 || input.L > 100.0 {
		errors = append(errors, validator.ValidationError{Field: "l", Message: "должно быть от 0.01 до 100.0"})
	}
	if input.MaterialID <= 0 {
		errors = append(errors, validator.ValidationError{Field: "material_id", Message: "обязательное поле"})
	}
	if input.Steps < 10 || input.Steps > 100000 {
		errors = append(errors, validator.ValidationError{Field: "steps", Message: "должно быть от 10 до 100000"})
	}
	if input.VuMin < 0 || input.VuMin > 100.0 {
		errors = append(errors, validator.ValidationError{Field: "vu_min", Message: "должно быть от 0 до 100.0"})
	}
	if input.VuMax < input.VuMin || input.VuMax > 100.0 {
		errors = append(errors, validator.ValidationError{Field: "vu_max", Message: "должно быть больше vu_min и не более 100.0"})
	}
	if input.VuSteps < 2 || input.VuSteps > 50 {
		errors = append(errors, validator.ValidationError{Field: "vu_steps", Message: "должно быть от 2 до 50"})
	}
	if input.TuMin < 0 || input.TuMin > 500.0 {
		errors = append(errors, validator.ValidationError{Field: "tu_min", Message: "должно быть от 0 до 500.0"})
	}
	if input.TuMax < input.TuMin || input.TuMax > 500.0 {
		errors = append(errors, validator.ValidationError{Field: "tu_max", Message: "должно быть больше tu_min и не более 500.0"})
	}
	if input.TuSteps < 2 || input.TuSteps > 50 {
		errors = append(errors, validator.ValidationError{Field: "tu_steps", Message: "должно быть от 2 до 50"})
	}

	return errors
}

func (s *CalculationService) Calculate(ctx context.Context, input *model.CalculationInput, userID int) (*model.CalculationResult, error) {
	startTime := time.Now()
	var startMem runtime.MemStats
	runtime.ReadMemStats(&startMem)

	errors := s.Validate(ctx, input)
	if len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	params, err := s.materialParamRepo.FindByMaterialID(ctx, input.MaterialID)
	if err != nil {
		return nil, err
	}

	result := s.compute(input, params)

	elapsed := time.Since(startTime)
	var endMem runtime.MemStats
	runtime.ReadMemStats(&endMem)

	memoryUsed := int(endMem.TotalAlloc - startMem.TotalAlloc)
	if memoryUsed < 0 {
		memoryUsed = 0
	}

	result.Metrics.CalcTimeMs = int(elapsed.Milliseconds())
	result.Metrics.MemoryUsedBytes = memoryUsed

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

func (s *CalculationService) CalculateSurface(ctx context.Context, input *model.CalculationSurfaceInput) (*model.CalculationSurfaceResult, error) {
	if errors := s.ValidateSurface(input); len(errors) > 0 {
		return nil, &ValidationError{Errors: errors}
	}

	params, err := s.materialParamRepo.FindByMaterialID(ctx, input.MaterialID)
	if err != nil {
		return nil, err
	}

	vuStep := (input.VuMax - input.VuMin) / float64(input.VuSteps-1)
	tuStep := (input.TuMax - input.TuMin) / float64(input.TuSteps-1)

	totalPoints := input.VuSteps * input.TuSteps

	type pointResult struct {
		Index int
		Point model.SurfacePoint
		Err   error
	}

	results := make(chan pointResult, totalPoints)

	numWorkers := runtime.NumCPU()
	if numWorkers > 8 {
		numWorkers = 8
	}

	var wg sync.WaitGroup

	type task struct {
		i, j   int
		vu, tu float64
	}

	tasks := make(chan task, totalPoints)

	for i := 0; i < input.VuSteps; i++ {
		vu := input.VuMin + float64(i)*vuStep
		for j := 0; j < input.TuSteps; j++ {
			tu := input.TuMin + float64(j)*tuStep
			tasks <- task{i: i, j: j, vu: vu, tu: tu}
		}
	}
	close(tasks)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				select {
				case <-ctx.Done():
					results <- pointResult{Err: ctx.Err()}
					return
				default:
					calcInput := &model.CalculationInput{
						W:          input.W,
						H:          input.H,
						L:          input.L,
						Vu:         task.vu,
						Tu:         task.tu,
						MaterialID: input.MaterialID,
						Steps:      input.Steps,
					}

					calcResult := s.compute(calcInput, params)

					results <- pointResult{
						Index: task.i*input.TuSteps + task.j,
						Point: model.SurfacePoint{
							Vu:           task.vu,
							Tu:           task.tu,
							Viscosity:    calcResult.Viscosity,
							Productivity: calcResult.Productivity,
						},
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	points := make([]model.SurfacePoint, totalPoints)
	for res := range results {
		if res.Err != nil {
			return nil, res.Err
		}
		points[res.Index] = res.Point
	}

	return &model.CalculationSurfaceResult{
		Points: points,
	}, nil
}

func (s *CalculationService) compute(input *model.CalculationInput, params map[string]float64) *model.CalculationResult {
	rho := getFloatParam(params, "density", 1380)
	c := getFloatParam(params, "heat_capacity", 2500)
	T0 := getFloatParam(params, "melting_temp", 145)
	mu0 := getFloatParam(params, "mu0", 12000)
	Ea := getFloatParam(params, "Ea", 147000)
	Tr := getFloatParam(params, "Tr", 180)
	n := getFloatParam(params, "n", 0.28)
	alphaU := getFloatParam(params, "alpha_u", 400)

	W := input.W
	H := input.H
	L := input.L
	Vu := input.Vu
	Tu := input.Tu
	steps := input.Steps

	if steps <= 0 {
		steps = 200
	}

	const R = 8.314

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
	dz := L / float64(steps)
	Qch := (W * H * Vu) / 2.0 * F
	productivity := rho * Qch * 3600

	profile := make([]model.Point, 0, steps+1)
	currentTemp := T0

	denominator := rho * c * Vu * H

	for i := 0; i <= steps; i++ {
		z := float64(i) * dz
		Tk := currentTemp + 273.15
		Trk := Tr + 273.15

		exponent := (Ea / R) * (1.0/Tk - 1.0/Trk)
		if exponent > 20 {
			exponent = 20
		}
		if exponent < -20 {
			exponent = -20
		}
		mu := mu0 * math.Exp(exponent)

		viscosity := mu * math.Pow(gammaDot, n-1)

		profile = append(profile, model.Point{
			X:           z,
			Temperature: currentTemp,
			Viscosity:   viscosity,
		})

		if i == steps {
			break
		}

		qGamma := viscosity * math.Pow(gammaDot, 2)
		qAlpha := alphaU * (Tu - currentTemp) / H
		dT := ((qGamma + qAlpha) / denominator) * dz

		maxDelta := 0.5
		if dT > maxDelta {
			dT = maxDelta
		}
		if dT < -maxDelta {
			dT = -maxDelta
		}

		currentTemp += dT

		if currentTemp < T0 {
			currentTemp = T0
		}
		if currentTemp > 250 {
			currentTemp = 250
		}
	}

	lastPoint := profile[len(profile)-1]

	return &model.CalculationResult{
		Productivity: productivity,
		Temperature:  lastPoint.Temperature,
		Viscosity:    lastPoint.Viscosity,
		Profile:      profile,
		Metrics: model.Metrics{
			CalcTimeMs:      0,
			MemoryUsedBytes: 0,
		},
	}
}

func getFloatParam(params map[string]float64, key string, defaultValue float64) float64 {
	if val, ok := params[key]; ok && val > 0 {
		return val
	}
	return defaultValue
}

type ValidationError struct {
	Errors []validator.ValidationError
}

func (e *ValidationError) Error() string {
	return "ошибка валидации"
}
