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
	opsCount := 0

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

	// Коэффициент формы канала
	var F float64
	if W > 0 && H > 0 {
		ratio := H / W
		opsCount++
		F = 0.125*math.Pow(ratio, 2) - 0.625*ratio + 1.0
		opsCount += 5
		if F < 0.3 {
			F = 0.3
		}
	} else {
		F = 1.0
	}

	// Параметр уравнения Андраде
	T0K := T0 + 273.0
	TrK := Tr + 273.0
	b := Ea / (R * (T0K + 20) * TrK)
	opsCount += 6

	// Объемный расход
	Qch := (W * H * Vu / 2.0) * F
	opsCount += 4

	// Скорость сдвига
	gammaDot := Vu / H
	opsCount++

	// Тепловые потоки
	qGamma := H * W * mu0 * math.Pow(gammaDot, n+1.0)
	opsCount += 5

	qAlpha := W * alphaU * Tu
	opsCount += 2

	// Знаменатель и коэффициент x1
	denom := W*(1.0+b*Tr)*alphaU - b*qAlpha
	opsCount += 6

	x1 := (b*qGamma + W*alphaU) / denom
	opsCount += 4

	// Шаг расчета
	dz := L / float64(steps)
	opsCount++

	roCQch := rho * c * Qch
	opsCount += 2

	profile := make([]model.Point, 0, steps+1)

	// Цикл по длине канала
	for i := 0; i <= steps; i++ {
		z := float64(i) * dz
		opsCount++

		expArg2 := -(denom * z) / roCQch
		x2 := 1.0 - math.Exp(expArg2)
		opsCount += 4

		tempNum := (W*(1.0/b+Tr)*alphaU - qAlpha) * z / roCQch
		expArg3 := b * (T0 - Tr - tempNum)
		x3 := math.Exp(expArg3)
		opsCount += 12

		xi := x1*x2 + x3
		opsCount += 2

		if xi <= 0 || math.IsNaN(xi) || math.IsInf(xi, 0) {
			xi = 1e-10
		}

		// Температура и вязкость
		temperature := Tr + (1.0/b)*math.Log(xi)
		opsCount += 3

		viscosity := mu0 * math.Exp(-b*(temperature-Tr)) * math.Pow(gammaDot, n-1.0)
		opsCount += 5

		profile = append(profile, model.Point{
			X:           z,
			Temperature: temperature,
			Viscosity:   viscosity,
		})
	}

	lastPoint := profile[len(profile)-1]

	// Производительность
	productivity := rho * Qch * 3600
	opsCount += 2

	return &model.CalculationResult{
		Productivity: productivity,
		Temperature:  lastPoint.Temperature,
		Viscosity:    lastPoint.Viscosity,
		Profile:      profile,
		Metrics: model.Metrics{
			CalcTimeMs:      0,
			MemoryUsedBytes: 0,
			OperationsCount: opsCount,
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
