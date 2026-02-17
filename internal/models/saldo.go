package models

import "time"

type SaldoTransaction struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id"`
	RoleID        uint      `json:"role_id"`
	TransactionNo string    `json:"transaction_no" gorm:"uniqueIndex"`
	Category      string    `json:"category"`
	Amount        int64     `json:"amount"`
	SaldoSetelah  int64     `json:"saldo_setelah"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

func (SaldoTransaction) TableName() string {
	return "myschema.saldo_role_transactions"
}
