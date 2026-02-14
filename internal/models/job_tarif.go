package models

import "time"

type JobTariff struct {
	ID               int64      `gorm:"primaryKey;column:id" json:"id"`
	JobCategoryID    int64      `gorm:"column:job_category_id" json:"job_category_id"`
	JobSubCategoryID *int64     `gorm:"column:job_sub_category_id" json:"job_sub_category_id,omitempty"`
	Price            float64    `gorm:"column:price" json:"price"`
	Unit             string     `gorm:"column:unit" json:"unit"`
	StartDate        time.Time  `gorm:"column:start_date" json:"start_date"`
	EndDate          *time.Time `gorm:"column:end_date" json:"end_date,omitempty"`
	IsActive         bool       `gorm:"column:is_active" json:"is_active"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (JobTariff) TableName() string {
	return "job_tariffs"
}
