package job_category

import (
	"teka-api/internal/models"

	"gorm.io/gorm"
)

type JobCategoryRepository interface {
	// JobCategory
	GetAll() ([]models.JobCategory, error)
	GetByID(id uint) (*models.JobCategory, error)
	Create(jobCategory *models.JobCategory) error
	Update(jobCategory *models.JobCategory) error
	Delete(id uint) error

	// JobSubCategory
	GetSubByID(id uint) (*models.JobSubCategory, error)
	GetSubAllByCategoryID(categoryID uint) ([]models.JobSubCategory, error)
	CreateSub(sub *models.JobSubCategory) error
	UpdateSub(sub *models.JobSubCategory) error
	DeleteSub(id uint) error
}

type jobCategoryRepo struct {
	db *gorm.DB
}

func NewJobCategoryRepository(db *gorm.DB) JobCategoryRepository {
	return &jobCategoryRepo{db}
}

// --- JobCategory ---
func (r *jobCategoryRepo) GetAll() ([]models.JobCategory, error) {
	var categories []models.JobCategory

	if err := r.db.
		Where("active = ?", true). // hanya active = true
		Order("id ASC").           // urutkan by id ascending
		Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *jobCategoryRepo) GetByID(id uint) (*models.JobCategory, error) {
	var category models.JobCategory
	if err := r.db.First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *jobCategoryRepo) Create(category *models.JobCategory) error {
	return r.db.Create(category).Error
}

func (r *jobCategoryRepo) Update(category *models.JobCategory) error {
	return r.db.Save(category).Error
}

func (r *jobCategoryRepo) Delete(id uint) error {
	return r.db.Delete(&models.JobCategory{}, id).Error
}

// --- JobSubCategory ---
func (r *jobCategoryRepo) GetSubByID(id uint) (*models.JobSubCategory, error) {
	var sub models.JobSubCategory
	if err := r.db.First(&sub, "id = ? AND active = ?", id, true).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *jobCategoryRepo) GetSubAllByCategoryID(categoryID uint) ([]models.JobSubCategory, error) {
	var subs []models.JobSubCategory
	if err := r.db.
		Where("job_category_id = ? AND active = ?", categoryID, true).
		Order("id ASC").
		Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *jobCategoryRepo) CreateSub(sub *models.JobSubCategory) error {
	return r.db.Create(sub).Error
}

func (r *jobCategoryRepo) UpdateSub(sub *models.JobSubCategory) error {
	return r.db.Save(sub).Error
}

func (r *jobCategoryRepo) DeleteSub(id uint) error {
	return r.db.Delete(&models.JobSubCategory{}, id).Error
}
