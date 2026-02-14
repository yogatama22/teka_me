package models

import "time"

type GlobalParameter struct {
	ID             int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ParameterCode  string    `gorm:"column:parameter_code;unique;not null" json:"parameter_code"`
	ParameterName  string    `gorm:"column:parameter_name" json:"parameter_name"`
	ParameterValue string    `gorm:"column:parameter_value" json:"parameter_value"`
	IsActive       bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedOn      time.Time `gorm:"column:created_on;autoCreateTime" json:"created_on"`
	CreatedBy      string    `gorm:"column:created_by" json:"created_by"`
	UpdatedOn      time.Time `gorm:"column:updated_on;autoUpdateTime" json:"updated_on"`
	UpdatedBy      string    `gorm:"column:updated_by" json:"updated_by"`
}

func (GlobalParameter) TableName() string {
	return "global_parameter"
}
