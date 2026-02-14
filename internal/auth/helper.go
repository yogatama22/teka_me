package auth

import (
	"errors"
	"fmt"
	"teka-api/internal/models"
	"teka-api/pkg/database"

	"gorm.io/gorm"
)

// GetUserWithActiveRole mengambil user berdasarkan ID
func GetUserWithActiveRole(userID uint) (UserResponse, error) {
	var user models.User

	err := database.DB.
		Preload("UserRoles", "active = ?", true).
		Preload("UserRoles.Role").
		First(&user, userID).Error
	if err != nil {
		return UserResponse{}, fmt.Errorf("user not found")
	}

	roleID := uint(0)
	roleName := ""
	if len(user.UserRoles) > 0 {
		roleID = user.UserRoles[0].Role.ID
		roleName = user.UserRoles[0].Role.Name
	}

	return UserResponse{
		ID:        user.ID,
		Nama:      user.Nama,
		Phone:     user.Phone,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		RoleID:    roleID,
		Roles:     roleName,
		Saldo:     0, // nanti di controller isi saldo
	}, nil
}

// GetSaldoByRole mengambil saldo terakhir berdasarkan user + role
func GetSaldoByRole(userID, roleID uint) (int64, error) {
	var saldo int64

	err := database.DB.
		Table("saldo_role_transactions").
		Select("saldo_setelah").
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Order("id DESC").
		Limit(1).
		Scan(&saldo).Error

	// Tangani jika record tidak ditemukan
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return saldo, nil
}
