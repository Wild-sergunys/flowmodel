package repository

import (
	"context"
	"database/sql"

	"flowmodel/internal/model"
)

type CalculationRepository interface {
	Create(ctx context.Context, calc *model.Calculation) error
	FindAll(ctx context.Context) ([]model.Calculation, error)
	FindByID(ctx context.Context, id int) (*model.Calculation, error)
}

type CalculationRepo struct {
	db *sql.DB
}

func NewCalculationRepo(db *sql.DB) *CalculationRepo {
	return &CalculationRepo{db: db}
}

func (r *CalculationRepo) Create(ctx context.Context, calc *model.Calculation) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO calculations (user_id, material_id, input_json, result_json) VALUES (?, ?, ?, ?)`,
		calc.UserID, calc.MaterialID, calc.InputJSON, calc.ResultJSON)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	calc.ID = int(id)
	return nil
}

func (r *CalculationRepo) FindAll(ctx context.Context) ([]model.Calculation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, material_id, input_json, result_json, created_at FROM calculations ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calcs []model.Calculation
	for rows.Next() {
		var c model.Calculation
		err := rows.Scan(&c.ID, &c.UserID, &c.MaterialID, &c.InputJSON, &c.ResultJSON, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		calcs = append(calcs, c)
	}
	return calcs, nil
}

func (r *CalculationRepo) FindByID(ctx context.Context, id int) (*model.Calculation, error) {
	var c model.Calculation
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, material_id, input_json, result_json, created_at FROM calculations WHERE id = ?`, id).
		Scan(&c.ID, &c.UserID, &c.MaterialID, &c.InputJSON, &c.ResultJSON, &c.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &c, nil
	}
}
