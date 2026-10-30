package model

import "time"

// PaymentStatus is the lifecycle state of a payment.
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "PENDING"
	PaymentSucceeded PaymentStatus = "SUCCEEDED"
	PaymentFailed    PaymentStatus = "FAILED"
	PaymentRefunded  PaymentStatus = "REFUNDED"
)

// Payment is a payment attempt for an order.
type Payment struct {
	ID         int64         `gorm:"primaryKey" json:"id"`
	OrderID    int64         `gorm:"index" json:"order_id"`
	Provider   string        `gorm:"size:32" json:"provider"`
	Amount     int64         `json:"amount"`
	Status     PaymentStatus `gorm:"size:32;index" json:"status"`
	ExternalID string        `gorm:"size:128" json:"external_id"`
	CreatedAt  time.Time     `json:"created_at"`
}
