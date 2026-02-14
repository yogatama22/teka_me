package job_tarif

import (
	"teka-api/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetByID(id int64) (*models.JobTariff, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(id int64) (*models.JobTariff, error) {
	var data models.JobTariff

	err := r.db.
		Where("id = ? AND is_active = true", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}
