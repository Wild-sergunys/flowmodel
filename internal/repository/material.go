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
