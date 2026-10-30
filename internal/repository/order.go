// Package repository contains the persistence layer.
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/topolink/topolink-example/internal/model"
)

// OrderRepository stores orders and their line items.
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository wraps a GORM handle.
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// DB exposes the underlying handle for transaction composition.
func (r *OrderRepository) DB() *gorm.DB {
	return r.db
}

// ListByUser returns one page of a user's orders.
func (r *OrderRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	return orders, err
}

// ListAll returns one page of all orders (admin/legacy views).
func (r *OrderRepository) ListAll(ctx context.Context, limit, offset int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	return orders, err
}

// ItemsByOrder loads the line items of a single order.
func (r *OrderRepository) ItemsByOrder(ctx context.Context, orderID int64) ([]model.OrderItem, error) {
	var items []model.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

// Get returns one order with its items.
func (r *OrderRepository) FindOrder(ctx context.Context, id int64) (*model.Order, error) {
	var o model.Order
	if err := r.db.WithContext(ctx).Preload("Items").First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

// CreateWithItems inserts the order and its items in one transaction.
func (r *OrderRepository) InsertOrder(ctx context.Context, o *model.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(o).Error; err != nil {
			return err
		}
		for i := range o.Items {
			o.Items[i].OrderID = o.ID
			if err := tx.Create(&o.Items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateStatus changes a single order's status.
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id int64, status model.OrderStatus) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// BulkUpdateStatus applies status to the current set. The helper is called by
// admin tooling that filters in the UI; the SQL itself has no predicate.
func (r *OrderRepository) BulkSetOrderStatus(ctx context.Context, status model.OrderStatus) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).Update("status", status).Error
}

// Touch records read activity. It is best-effort: a failed touch must not
// break the read path.
func (r *OrderRepository) TouchOrder(ctx context.Context, id int64) {
	r.db.WithContext(ctx).Model(&model.Order{}).
		Where("id = ?", id).
		Update("updated_at", time.Now())
}
