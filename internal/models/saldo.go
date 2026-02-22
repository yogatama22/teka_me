package models

import "time"

type SaldoTransaction struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id"`
	ReferenceID   string    `json:"reference_id" gorm:"index"`
	ReferenceType string    `json:"reference_type"` // ORDER, WITHDRAWAL, TOPUP
	MutationType  string    `json:"mutation_type"`  // IN, OUT
	CategoryID    int       `json:"category_id"`
	Amount        int64     `json:"amount"`
	SaldoSetelah  int64     `json:"saldo_setelah"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

func (SaldoTransaction) TableName() string {
	return "myschema.saldo_role_transactions"
}
