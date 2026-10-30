package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/topolink/topolink-example/internal/model"
)

// LedgerRepository stores balances and double-entry records.
type LedgerRepository struct {
	db *gorm.DB
}

// NewLedgerRepository wraps a GORM handle.
func NewLedgerRepository(db *gorm.DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

// Debit subtracts amount from an account and records the entry.
func (r *LedgerRepository) PostDebit(ctx context.Context, tx *gorm.DB, accountID, orderID, amount int64) error {
	if err := tx.WithContext(ctx).Table("accounts").
		Where("id = ?", accountID).
		Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
		return err
	}
	entry := model.LedgerEntry{AccountID: accountID, OrderID: orderID, Amount: amount, Direction: "debit"}
	return tx.WithContext(ctx).Create(&entry).Error
}

// Credit adds amount to an account and records the entry.
func (r *LedgerRepository) PostCredit(ctx context.Context, tx *gorm.DB, accountID, orderID, amount int64) error {
	if err := tx.WithContext(ctx).Table("accounts").
		Where("id = ?", accountID).
		Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
		return err
	}
	entry := model.LedgerEntry{AccountID: accountID, OrderID: orderID, Amount: amount, Direction: "credit"}
	return tx.WithContext(ctx).Create(&entry).Error
}

// MarkSettled records the result of a settlement batch.
func (r *LedgerRepository) MarkSettlementSettled(ctx context.Context, batchID string, amount int64) error {
	return r.db.WithContext(ctx).Table("settlements").
		Where("batch_id = ?", batchID).
		Updates(map[string]any{"status": "SETTLED", "amount": amount}).Error
}
