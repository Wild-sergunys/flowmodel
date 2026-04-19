package model

import "time"

type User struct {
	ID int `json:"id"`
	Login string `json:"login"`
	PasswordHash string `json:"-"` // не отдаем в JSON
	Role string `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}