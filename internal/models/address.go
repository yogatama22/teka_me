package models

import (
	"time"
)

type AddressType struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

type UserAddress struct {
	ID         uint        `gorm:"primaryKey"`
	UserID     uint        `gorm:"not null"`
	User       User        `gorm:"foreignKey:UserID"`
	TypeID     uint        `gorm:"not null"`
	Type       AddressType `gorm:"foreignKey:TypeID"`
	Address    string      `gorm:"not null"`
	City       string
	PostalCode string
	Phone      string
	IsPrimary  bool      `gorm:"default:false"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

type UserAddressResponse struct {
	ID         uint   `json:"id"`
	TypeID     uint   `json:"type_id"`
	TypeName   string `json:"type_name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Phone      string `json:"phone"`
	IsPrimary  bool   `json:"is_primary"`
}

func MapAddressToResponse(addr UserAddress) UserAddressResponse {
	return UserAddressResponse{
		ID:         addr.ID,
		TypeID:     addr.TypeID,
		TypeName:   addr.Type.Name,
		Address:    addr.Address,
		City:       addr.City,
		PostalCode: addr.PostalCode,
		Phone:      addr.Phone,
		IsPrimary:  addr.IsPrimary,
	}
}
