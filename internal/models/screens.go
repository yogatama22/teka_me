package models

import "time"

type Screen struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ImageURL  string    `json:"image_url"`
	Status    bool      `json:"status"`
	OrderBy   int       `json:"order_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
