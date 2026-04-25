package validator

import (
	"context"
	"testing"

	"flowmodel/internal/model"
	"flowmodel/internal/repository"
)

type mockMaterialRepo struct {
	materials map[int]*model.Material
	err       error
}

func (m *mockMaterialRepo) FindAll(ctx context.Context) ([]model.Material, error) {
	return nil, nil
}

func (m *mockMaterialRepo) FindByID(ctx context.Context, id int) (*model.Material, error) {
	if m.err != nil {
		return nil, m.err
	}
	mat, ok := m.materials[id]
	if !ok {
		return nil, nil
	}
	return mat, nil
}

func (m *mockMaterialRepo) Create(ctx context.Context, mat *model.Material) error {
	return nil
}

func (m *mockMaterialRepo) Update(ctx context.Context, mat *model.Material) error {
	return nil
}

func (m *mockMaterialRepo) Delete(ctx context.Context, id int) error {
	return nil
}

// Проверка, что мок реализует интерфейс
var _ repository.MaterialRepository = (*mockMaterialRepo)(nil)

func TestValidateCalculationInput(t *testing.T) {
	repo := &mockMaterialRepo{
		materials: map[int]*model.Material{
			1: {ID: 1, Name: "ПВХ"},
		},
	}

	tests := []struct {
		name       string
		input      model.CalculationInput
		wantCount  int
		wantFields []string
	}{
		// Пустые поля
		{
			name:       "все поля пустые",
			input:      model.CalculationInput{},
			wantCount:  5,
			wantFields: []string{"w", "h", "l", "material_id", "steps"},
		},
		// Нормальные значения
		{
			name: "нормальный ввод",
			input: model.CalculationInput{
				W: 0.25, H: 0.01, L: 9.5,
				Vu: 1.5, Tu: 150,
				MaterialID: 1,
				Steps:      100,
			},
			wantCount: 0,
		},
		// Границы
		{
			name: "w меньше минимума",
			input: model.CalculationInput{
				W: 0.0001, H: 0.01, L: 1.0,
				Vu: 1, Tu: 100,
				MaterialID: 1,
				Steps:      10,
			},
			wantCount:  1,
			wantFields: []string{"w"},
		},
		{
			name: "w больше максимума",
			input: model.CalculationInput{
				W: 11.0, H: 0.01, L: 1.0,
				Vu: 1, Tu: 100,
				MaterialID: 1,
				Steps:      10,
			},
			wantCount:  1,
			wantFields: []string{"w"},
		},
		{
			name: "steps меньше минимума",
			input: model.CalculationInput{
				W: 0.25, H: 0.01, L: 1.0,
				Vu: 1, Tu: 100,
				MaterialID: 1,
				Steps:      5,
			},
			wantCount:  1,
			wantFields: []string{"steps"},
		},
		// Материал не найден
		{
			name: "материал не существует",
			input: model.CalculationInput{
				W: 0.25, H: 0.01, L: 1.0,
				Vu: 1, Tu: 100,
				MaterialID: 999,
				Steps:      10,
			},
			wantCount:  1,
			wantFields: []string{"material_id"},
		},
		// Несколько ошибок
		{
			name: "две ошибки валидации",
			input: model.CalculationInput{
				W: 0.0001, H: 0.0001, L: 1.0,
				Vu: 1, Tu: 100,
				MaterialID: 1,
				Steps:      10,
			},
			wantCount:  2,
			wantFields: []string{"w", "h"},
		},
		// Граничные значения (должны пройти)
		{
			name: "все поля на нижних границах",
			input: model.CalculationInput{
				W: 0.001, H: 0.001, L: 0.01,
				Vu: 0, Tu: 0,
				MaterialID: 1,
				Steps:      10,
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateCalculationInput(&tt.input, repo)

			if len(errors) != tt.wantCount {
				t.Errorf("ожидалось %d ошибок, получено %d: %+v", tt.wantCount, len(errors), errors)
			}

			if tt.wantFields != nil {
				gotFields := make(map[string]bool)
				for _, e := range errors {
					gotFields[e.Field] = true
				}
				for _, wantField := range tt.wantFields {
					if !gotFields[wantField] {
						t.Errorf("ожидалась ошибка в поле %q, но её нет среди: %+v", wantField, errors)
					}
				}
			}
		})
	}
}
