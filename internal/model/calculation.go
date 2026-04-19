package model

type CalculationInput struct {
	// Геометрия
	W float64 `json:"w"`
	H float64 `json:"h"`
	L float64 `json:"l"`

	// Режимные параметры
	Vu float64 `json:"vu"`
	Tu float64 `json:"tu"`

	// Материал
	MaterialID int     `json:"material_id"`
	T0         float64 `json:"t0"`

	// Расчёт
	Steps int `json:"steps"`
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
	Productivity float64 `json:"productivity"`
	Temperature  float64 `json:"temperature"`
	Viscosity    float64 `json:"viscosity"`
	Profile      []Point `json:"profile"`
	Metrics      Metrics `json:"metrics"`
}
