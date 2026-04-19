package repository

import (
	"context"
	"database/sql"
	"flowmodel/internal/model"
)

// >>>> СТРУКТУРЫ <<<<

type MaterialRepo struct {
	db *sql.DB
}

// >>>> ИНТЕРФЕЙСЫ <<<<

type MaterialRepository interface {
	FindAll(ctx context.Context) ([]model.Material, error)
	FindByID(ctx context.Context, id int) (*model.Material, error)
	Create(ctx context.Context, m *model.Material) error
	Update(ctx context.Context, m *model.Material) error
	Delete(ctx context.Context, id int) error
}

// >>>> ФУНКЦИИ И МЕТОДЫ <<<<

func NewMaterialRepo(db *sql.DB) *MaterialRepo {
	return &MaterialRepo{db: db}
}

func (r *MaterialRepo) FindAll(ctx context.Context) ([]model.Material, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, description, created_at, updated_at FROM materials")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []model.Material
	for rows.Next() {
		var m model.Material
		err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		materials = append(materials, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return materials, nil
}

func (r *MaterialRepo) FindByID(ctx context.Context, id int) (*model.Material, error) {
	var m model.Material
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM materials WHERE id = ?`, id).
		Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt, &m.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &m, nil
	}
}

func (r *MaterialRepo) Create(ctx context.Context, m *model.Material) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO materials (name, description) VALUES (?, ?)`,
		m.Name, m.Description)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	m.ID = int(id)
	return nil
}

func (r *MaterialRepo) Update(ctx context.Context, m *model.Material) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE materials SET name=?, description=? WHERE id=?`,
		m.Name, m.Description, m.ID)
	return err
}

func (r *MaterialRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM materials WHERE id=?`, id)
	return err
}
