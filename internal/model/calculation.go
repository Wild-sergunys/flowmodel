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
