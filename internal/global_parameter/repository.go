package global_parameter

import (
	"context"
	"teka-api/internal/models"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// GetAll ambil semua global parameter
func (r *Repository) GetAll(ctx context.Context) ([]models.GlobalParameter, error) {
	var params []models.GlobalParameter
	err := r.DB.WithContext(ctx).Find(&params).Error
	return params, err
}

// GetByID
func (r *Repository) GetByID(ctx context.Context, id int) (*models.GlobalParameter, error) {
	var param models.GlobalParameter
	err := r.DB.WithContext(ctx).First(&param, id).Error
	if err != nil {
		return nil, err
	}
	return &param, nil
}

// Create
func (r *Repository) Create(ctx context.Context, param *models.GlobalParameter) error {
	return r.DB.WithContext(ctx).Create(param).Error
}

// Update
func (r *Repository) Update(ctx context.Context, param *models.GlobalParameter) error {
	return r.DB.WithContext(ctx).Save(param).Error
}

// Delete
func (r *Repository) Delete(ctx context.Context, id int) error {
	return r.DB.WithContext(ctx).Delete(&models.GlobalParameter{}, id).Error
}
