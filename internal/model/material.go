package model

import (
	"time"
)

// Material представляет тип прерабатываемого материала
// Description - (*string) + тег omitempty
// Для различия "поле не прислали" (nil, в JSON нет)
// от "прислали пустую строку" (&"", в JSON как "")
type Material struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
