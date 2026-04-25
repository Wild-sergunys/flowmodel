package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/Wild-sergunys/flowmodel/internal/model"
)

type CalculationRepository interface {
	Create(ctx context.Context, calc *model.Calculation) error
	FindAll(ctx context.Context) ([]model.Calculation, error)
	FindByID(ctx context.Context, id int) (*model.Calculation, error)
	FindByUserID(ctx context.Context, userID int) ([]model.Calculation, error)
}

type CalculationRepo struct {
	db *sql.DB
}

func NewCalculationRepo(db *sql.DB) *CalculationRepo {
	return &CalculationRepo{db: db}
}

func (r *CalculationRepo) Create(ctx context.Context, calc *model.Calculation) error {
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		result, err := r.db.ExecContext(ctx,
			`INSERT INTO calculations (user_id, material_id, input_json, result_json) VALUES (?, ?, ?, ?)`,
			calc.UserID, calc.MaterialID, calc.InputJSON, calc.ResultJSON)

		if err == nil {
			id, _ := result.LastInsertId()
			calc.ID = int(id)
			return nil
		}

		lastErr = err

		// Если ошибка не про соединение с БД - нет смысла ретраить
		if !isRetryableError(err) {
			return err
		}

		if attempt < 2 {
			time.Sleep(time.Millisecond * time.Duration(10*(attempt+1)))
		}
	}

	return lastErr
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

func (r *CalculationRepo) FindByUserID(ctx context.Context, userID int) ([]model.Calculation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, material_id, input_json, result_json, created_at 
        FROM calculations 
        WHERE user_id = ? 
        ORDER BY created_at DESC`, userID)
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

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Таймауты, потеря соединения, дедлоки - ретраим
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"too many connections",
		"deadlock",
		"lock wait timeout",
		"try restarting transaction",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}
