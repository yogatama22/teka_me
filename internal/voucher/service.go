package voucher

import "teka-api/internal/models"

type Service interface {
	GetActiveVouchers() ([]models.Voucher, error)
	GetAvailableVouchers(userID uint) ([]models.Voucher, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetActiveVouchers() ([]models.Voucher, error) {
	return s.repo.GetActiveVouchers()
}

func (s *service) GetAvailableVouchers(userID uint) ([]models.Voucher, error) {
	return s.repo.GetAvailableVouchers(userID)
}
