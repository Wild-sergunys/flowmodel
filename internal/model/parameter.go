package model

import "time"

type Parameter struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Unit        *string   `json:"unit,omitempty"`
	DataType    string    `json:"data_type"`
	Category    string    `json:"category"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
