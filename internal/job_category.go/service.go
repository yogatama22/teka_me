package job_category

import (
	"teka-api/internal/models"
)

type JobCategoryService interface {
	// JobCategory

	GetAll() ([]models.JobCategory, error)
	GetByID(id uint) (*models.JobCategory, error)
	Create(category *models.JobCategory) error
	Update(category *models.JobCategory) error
	Delete(id uint) error

	// JobSubCategory
	GetSubByID(id uint) (*models.JobSubCategory, error)
	GetSubAllByCategoryID(categoryID uint) ([]models.JobSubCategory, error)
	CreateSub(sub *models.JobSubCategory) error
	UpdateSub(sub *models.JobSubCategory) error
	DeleteSub(id uint) error
}

type jobCategoryService struct {
	repo JobCategoryRepository
}

func NewJobCategoryService(repo JobCategoryRepository) JobCategoryService {
	return &jobCategoryService{repo}
}

// --- JobCategory ---
func (s *jobCategoryService) GetAll() ([]models.JobCategory, error) {
	return s.repo.GetAll()
}

func (s *jobCategoryService) GetByID(id uint) (*models.JobCategory, error) {
	return s.repo.GetByID(id)
}

func (s *jobCategoryService) Create(category *models.JobCategory) error {
	return s.repo.Create(category)
}

func (s *jobCategoryService) Update(category *models.JobCategory) error {
	return s.repo.Update(category)
}

func (s *jobCategoryService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// --- JobSubCategory ---
func (s *jobCategoryService) GetSubByID(id uint) (*models.JobSubCategory, error) {
	return s.repo.GetSubByID(id)
}

func (s *jobCategoryService) GetSubAllByCategoryID(categoryID uint) ([]models.JobSubCategory, error) {
	return s.repo.GetSubAllByCategoryID(categoryID)
}

func (s *jobCategoryService) CreateSub(sub *models.JobSubCategory) error {
	return s.repo.CreateSub(sub)
}

func (s *jobCategoryService) UpdateSub(sub *models.JobSubCategory) error {
	return s.repo.UpdateSub(sub)
}

func (s *jobCategoryService) DeleteSub(id uint) error {
	return s.repo.DeleteSub(id)
}
