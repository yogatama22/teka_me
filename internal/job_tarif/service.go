package job_tarif

import (
	"teka-api/internal/models"
)

type Service interface {
	GetByID(id int64) (*models.JobTariff, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(id int64) (*models.JobTariff, error) {
	return s.repo.GetByID(id)
}
