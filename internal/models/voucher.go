package models

import "time"

type Voucher struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	DiscountType  string    `json:"discount_type"` // PERCENT | FIXED
	DiscountValue float64   `json:"discount_value"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	IsActive      bool      `json:"is_active"`
	MaxUsage      *int      `json:"max_usage"`
	UsedCount     int       `json:"used_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Voucher) TableName() string {
	return "vouchers"
}
