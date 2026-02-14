package screens

import (
	"context"

	"teka-api/internal/models"
	"teka-api/pkg/database"

	"gorm.io/gorm"
)

type Repository interface {
	GetAll(ctx context.Context) ([]models.Screen, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll(ctx context.Context) ([]models.Screen, error) {
	var screens []models.Screen

	err := r.db.
		WithContext(ctx).
		Table(database.Table("screens")). // 👈 BACA DARI ENV
		Where("status = ?", true).
		Order("order_by ASC").
		Find(&screens).Error

	if err != nil {
		return nil, err
	}

	return screens, nil
}
