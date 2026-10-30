package model

import "time"

// OrderStatus is the lifecycle state of an order.
type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusPaid      OrderStatus = "PAID"
	StatusShipped   OrderStatus = "SHIPPED"
	StatusCancelled OrderStatus = "CANCELLED"
	StatusRefunded  OrderStatus = "REFUNDED"
)

// Order is a customer order.
type Order struct {
	ID        int64       `gorm:"primaryKey" json:"id"`
	UserID    int64       `gorm:"index" json:"user_id"`
	Status    OrderStatus `gorm:"size:32;index" json:"status"`
	Currency  string      `gorm:"size:8" json:"currency"`
	Total     int64       `json:"total"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Items     []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

// OrderItem is one line item.
type OrderItem struct {
	ID        int64  `gorm:"primaryKey" json:"id"`
	OrderID   int64  `gorm:"index" json:"order_id"`
	SKU       string `gorm:"size:64" json:"sku"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

// Subtotal returns the line total in cents.
func (i OrderItem) Subtotal() int64 {
	return int64(i.Quantity) * i.UnitPrice
}

// CanTransitionTo reports whether the order may move to next.
func (o Order) CanTransitionTo(next OrderStatus) bool {
	switch o.Status {
	case StatusPending:
		return next == StatusPaid || next == StatusCancelled
	case StatusPaid:
		return next == StatusShipped || next == StatusRefunded
	case StatusShipped:
		return next == StatusRefunded
	default:
		return false
	}
}
