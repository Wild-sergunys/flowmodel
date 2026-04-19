package repository

import (
	"context"
	"database/sql"

	"flowmodel/internal/model"
)

// >>>> СТРУКТУРЫ <<<<

type UserRepo struct {
	db *sql.DB
}

// >>>> ИНТЕРФЕЙСЫ <<<<

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByLogin(ctx context.Context, login string) (*model.User, error)
}

// >>>> ФУНКЦИИ И МЕТОДЫ <<<<

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (login, password_hash, role) VALUES (?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, user.Login, user.PasswordHash, user.Role)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)
	return nil
}


func (r *UserRepo) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `SELECT id, login, password_hash, role, created_at, updated_at FROM users WHERE login = ?`
	row := r.db.QueryRowContext(ctx, query, login)


	var user model.User
	err := row.Scan(&user.ID, &user.Login, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &user, nil
	}
}