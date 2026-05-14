package model

import "time"

type CalculationInput struct {
	// Геометрия
	W float64 `json:"w"`
	H float64 `json:"h"`
	L float64 `json:"l"`

	// Режимные параметры
	Vu float64 `json:"vu"`
	Tu float64 `json:"tu"`

	// Материал
	MaterialID int `json:"material_id"`

	// Расчёт
	Steps int `json:"steps"`
}

type CalculationSurfaceInput struct {
	// Геометрия (фиксирована для всей поверхности)
	W float64 `json:"w"`
	H float64 `json:"h"`
	L float64 `json:"l"`

	// Материал
	MaterialID int `json:"material_id"`

	// Точность расчёта для каждой точки
	Steps int `json:"steps"`

	// Диапазон скоростей крышки
	VuMin   float64 `json:"vu_min"`
	VuMax   float64 `json:"vu_max"`
	VuSteps int     `json:"vu_steps"`

	// Диапазон температур крышки
	TuMin   float64 `json:"tu_min"`
	TuMax   float64 `json:"tu_max"`
	TuSteps int     `json:"tu_steps"`
}

type Point struct {
	X           float64 `json:"x"`
	Temperature float64 `json:"temperature"`
	Viscosity   float64 `json:"viscosity"`
}

type Metrics struct {
	CalcTimeMs      int `json:"calc_time_ms"`
	MemoryUsedBytes int `json:"memory_used_bytes"`
}

type CalculationResult struct {
	ID           int     `json:"id,omitempty"`
	Productivity float64 `json:"productivity"`
	Temperature  float64 `json:"temperature"`
	Viscosity    float64 `json:"viscosity"`
	Profile      []Point `json:"profile"`
	Metrics      Metrics `json:"metrics"`
}

type SurfacePoint struct {
	Vu           float64 `json:"vu"`
	Tu           float64 `json:"tu"`
	Viscosity    float64 `json:"viscosity"`
	Productivity float64 `json:"productivity"`
}

type CalculationSurfaceResult struct {
	Points []SurfacePoint `json:"points"`
}

type Calculation struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	MaterialID int       `json:"material_id"`
	InputJSON  string    `json:"input_json"`
	ResultJSON string    `json:"result_json"`
	CreatedAt  time.Time `json:"created_at"`
}
