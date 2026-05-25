package repository

import (
	"context"
	"database/sql"

	"github.com/Wild-sergunys/flowmodel/internal/model"
)

type MaterialParameterRepository interface {
	FindByMaterialID(ctx context.Context, materialID int) (map[string]float64, error)
	FindDetailsByMaterialID(ctx context.Context, materialID int) ([]model.MaterialParameterValue, error)
	Update(ctx context.Context, materialID int, params map[string]float64) error // ← новое
}

type MaterialParameterRepo struct {
	db *sql.DB
}

func NewMaterialParameterRepo(db *sql.DB) *MaterialParameterRepo {
	return &MaterialParameterRepo{db: db}
}

func (r *MaterialParameterRepo) FindByMaterialID(ctx context.Context, materialID int) (map[string]float64, error) {
	query := `
        SELECT p.code, mp.value_float 
        FROM material_parameters mp
        JOIN parameters p ON mp.parameter_id = p.id
        WHERE mp.material_id = ? AND p.data_type = 'float'
    `
	rows, err := r.db.QueryContext(ctx, query, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	params := make(map[string]float64)
	for rows.Next() {
		var code string
		var value float64
		if err := rows.Scan(&code, &value); err != nil {
			return nil, err
		}
		params[code] = value
	}
	return params, nil
}

func (r *MaterialParameterRepo) FindDetailsByMaterialID(
	ctx context.Context,
	materialID int,
) ([]model.MaterialParameterValue, error) {
	query := `
        SELECT p.code, p.name, p.unit, p.category, p.description, mp.value_float
        FROM material_parameters mp
        JOIN parameters p ON mp.parameter_id = p.id
        WHERE mp.material_id = ? AND p.data_type = 'float'
        ORDER BY p.id
    `
	rows, err := r.db.QueryContext(ctx, query, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	params := make([]model.MaterialParameterValue, 0)
	for rows.Next() {
		var p model.MaterialParameterValue
		if err := rows.Scan(&p.Code, &p.Name, &p.Unit, &p.Category, &p.Description, &p.ValueFloat); err != nil {
			return nil, err
		}
		params = append(params, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return params, nil
}

func (r *MaterialParameterRepo) Update(ctx context.Context, materialID int, params map[string]float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM material_parameters WHERE material_id = ?`, materialID)
	if err != nil {
		return err
	}

	for code, value := range params {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO material_parameters (material_id, parameter_id, value_float)
			SELECT ?, p.id, ?
			FROM parameters p
			WHERE p.code = ?
		`, materialID, value, code)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
