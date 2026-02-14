package models

import "time"

type JobCategory struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"size:50;not null;unique"`
	Active    *bool     `json:"active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type JobSubCategory struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	JobCategoryID   uint      `json:"job_category_id" gorm:"not null"`
	SubCategoryName string    `json:"sub_category_name" gorm:"size:100;not null"`
	Description     *string   `json:"description"`
	Active          *bool     `json:"active" gorm:"default:false"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
