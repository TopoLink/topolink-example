package model

import "time"

// Account is a user's balance account.
type Account struct {
	ID       int64  `gorm:"primaryKey" json:"id"`
	UserID   int64  `gorm:"index" json:"user_id"`
	Balance  int64  `json:"balance"`
	Currency string `gorm:"size:8" json:"currency"`
}

// LedgerEntry is an immutable double-entry record.
type LedgerEntry struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	AccountID int64     `gorm:"index" json:"account_id"`
	OrderID   int64     `gorm:"index" json:"order_id"`
	Amount    int64     `json:"amount"`
	Direction string    `gorm:"size:8" json:"direction"`
	CreatedAt time.Time `json:"created_at"`
}

// Settlement is a daily settlement batch.
type Settlement struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	BatchID   string    `gorm:"size:64;uniqueIndex" json:"batch_id"`
	Amount    int64     `json:"amount"`
	Status    string    `gorm:"size:32" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
