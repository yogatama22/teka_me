package auth

import "time"

type UserResponse struct {
	ID        uint      `json:"id"`
	Nama      string    `json:"nama"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uint      `json:"role_id"`
	Roles     string    `json:"roles"` // <- ubah json tag dari role_name → roles
	Saldo     int64     `json:"saldo"`
}
