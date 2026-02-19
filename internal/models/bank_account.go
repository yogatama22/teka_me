package models

type WithdrawRequest struct {
	Amount        int64  `json:"amount"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	AccountHolder string `json:"account_holder"`
}
