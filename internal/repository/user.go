package repository

import (
	"context"
	"database/sql"

	"github.com/Wild-sergunys/flowmodel/internal/model"
)

// >>>> СТРУКТУРЫ <<<<

type UserRepo struct {
	db *sql.DB
}

// >>>> ИНТЕРФЕЙСЫ <<<<

// Дополнить интерфейс UserRepository
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByLogin(ctx context.Context, login string) (*model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int) error
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

func (r *UserRepo) FindAll(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, login, role, created_at, updated_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.ID, &u.Login, &u.Role, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id int) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx, `SELECT id, login, role, created_at, updated_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Login, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &u, nil
	}
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role=? WHERE id=?`, user.Role, user.ID)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id=?`, id)
	return err
}
