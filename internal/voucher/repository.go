package voucher

import (
	"teka-api/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetActiveVouchers() ([]models.Voucher, error)
	GetAvailableVouchers(userID uint) ([]models.Voucher, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ADMIN
func (r *repository) GetActiveVouchers() ([]models.Voucher, error) {
	var vouchers []models.Voucher

	err := r.db.
		Where("is_active = true").
		Where("CURRENT_DATE BETWEEN start_date AND end_date").
		Where("(max_usage IS NULL OR used_count < max_usage)").
		Find(&vouchers).Error

	return vouchers, err
}

// USER
func (r *repository) GetAvailableVouchers(userID uint) ([]models.Voucher, error) {
	var vouchers []models.Voucher

	err := r.db.
		Where("is_active = true").
		Where("CURRENT_DATE BETWEEN start_date AND end_date").
		Where("(max_usage IS NULL OR used_count < max_usage)").
		Where(`
			NOT EXISTS (
				SELECT 1
				FROM voucher_histories vh
				WHERE vh.voucher_id = vouchers.id
				  AND vh.user_id = ?
			)
		`, userID).
		Find(&vouchers).Error

	return vouchers, err
}
