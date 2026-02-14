package global_parameter

import (
	"context"
	"teka-api/internal/models"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll(ctx context.Context) ([]models.GlobalParameter, error) {
	return s.Repo.GetAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (*models.GlobalParameter, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, param *models.GlobalParameter) error {
	return s.Repo.Create(ctx, param)
}

func (s *Service) Update(ctx context.Context, param *models.GlobalParameter) error {
	return s.Repo.Update(ctx, param)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.Repo.Delete(ctx, id)
}
