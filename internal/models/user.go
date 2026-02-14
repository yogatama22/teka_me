package models

import (
	"teka-api/pkg/database"
	"time"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Nama      string `gorm:"not null"`
	Phone     string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Pin       string
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relasi ke pivot, termasuk info Active
	UserRoles []UserRole `gorm:"foreignKey:UserID"`
}

// Pivot table dengan relasi ke Role
type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
	Active bool `gorm:"default:false"`

	Role Role `gorm:"foreignKey:RoleID"` // relasi ke Role
}

type TempUser struct {
	ID        uint `gorm:"primaryKey"`
	Nama      string
	Phone     string
	Email     string
	Password  string
	CreatedAt time.Time
}

type OTP struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"not null"`
	Otp       string    `gorm:"not null"`
	ExpiredAt time.Time `gorm:"not null"`
}

type Role struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

// SCHEMA TABLE \\
func (User) TableName() string {
	return database.Table("users")
}

func (UserRole) TableName() string {
	return database.Table("user_roles")
}

func (TempUser) TableName() string {
	return database.Table("temp_users")
}

func (OTP) TableName() string {
	return database.Table("otps")
}

func (Role) TableName() string {
	return database.Table("roles")
}
