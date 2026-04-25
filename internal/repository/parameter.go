package repository

import (
	"context"
	"database/sql"

	"github.com/Wild-sergunys/flowmodel/internal/model"
)

// >>>> СТРУКТУРЫ <<<<

type ParameterRepo struct {
	db *sql.DB
}

// >>>> ИНТЕРФЕЙСЫ <<<<

type ParameterRepository interface {
	FindAll(ctx context.Context) ([]model.Parameter, error)
	FindByID(ctx context.Context, id int) (*model.Parameter, error)
	Create(ctx context.Context, p *model.Parameter) error
	Update(ctx context.Context, p *model.Parameter) error
	Delete(ctx context.Context, id int) error
}

// >>>> ФУНКЦИИ И МЕТОДЫ <<<<

func NewParameterRepo(db *sql.DB) *ParameterRepo {
	return &ParameterRepo{db: db}
}

func (r *ParameterRepo) FindAll(ctx context.Context) ([]model.Parameter, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, code, name, unit, data_type, category, description, created_at FROM parameters`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var params []model.Parameter
	for rows.Next() {
		var p model.Parameter
		err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Unit, &p.DataType, &p.Category, &p.Description, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		params = append(params, p)
	}
	return params, nil
}

func (r *ParameterRepo) FindByID(ctx context.Context, id int) (*model.Parameter, error) {
	var p model.Parameter
	err := r.db.QueryRowContext(ctx, `SELECT id, code, name, unit, data_type, category, description, created_at FROM parameters WHERE id = ?`, id).
		Scan(&p.ID, &p.Code, &p.Name, &p.Unit, &p.DataType, &p.Category, &p.Description, &p.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &p, nil
	}
}

func (r *ParameterRepo) Create(ctx context.Context, p *model.Parameter) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO parameters (code, name, unit, data_type, category,description) VALUES (?, ?, ?, ?, ?, ?)`,
		p.Code, p.Name, p.Unit, p.DataType, p.Category, p.Description)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	p.ID = int(id)
	return nil
}

func (r *ParameterRepo) Update(ctx context.Context, p *model.Parameter) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE parameters SET
		code=?, name=?, unit=?, data_type=?, category=?, description=? WHERE id=?`,
		p.Code, p.Name, p.Unit, p.DataType, p.Category, p.Description, p.ID)
	return err
}

func (r *ParameterRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM parameters WHERE id=?`, id)
	return err
}
