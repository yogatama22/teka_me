package screens

import (
	"context"

	"teka-api/internal/models"
)

type Service interface {
	GetAll(ctx context.Context) ([]models.Screen, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAll(ctx context.Context) ([]models.Screen, error) {
	return s.repo.GetAll(ctx)
}
