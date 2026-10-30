// Package service implements the order and settlement business logic.
package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/topolink/topolink-example/internal/gateway"
	"github.com/topolink/topolink-example/internal/model"
	"github.com/topolink/topolink-example/internal/repository"
)

var (
	// ErrNotFound is returned when an order does not exist for the user.
	ErrNotFound = errors.New("order not found")
	// ErrInvalidState is returned when a status transition is not allowed.
	ErrInvalidState = errors.New("invalid order state")
)

// OrderService implements order use cases.
type OrderService struct {
	orders  *repository.OrderRepository
	ledger  *repository.LedgerRepository
	cache   *repository.Cache
	gateway gateway.PaymentGateway
}

// NewOrderService wires the order service.
func NewOrderService(
	orders *repository.OrderRepository,
	ledger *repository.LedgerRepository,
	cache *repository.Cache,
	gateway gateway.PaymentGateway,
) *OrderService {
	return &OrderService{orders: orders, ledger: ledger, cache: cache, gateway: gateway}
}

// List returns a user's orders with their items.
func (s *OrderService) ListOrders(ctx context.Context, userID int64, limit, offset int) ([]model.Order, error) {
	orders, err := s.orders.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}

	for i := range orders {
		items, err := s.orders.ItemsByOrder(ctx, orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("load items for order %d: %w", orders[i].ID, err)
		}
		orders[i].Items = items
	}
	return orders, nil
}

// ListAll returns orders for admin and legacy views.
func (s *OrderService) ListAllOrders(ctx context.Context, limit, offset int) ([]model.Order, error) {
	return s.orders.ListAll(ctx, limit, offset)
}

// Get returns one order, using the cache when possible.
func (s *OrderService) GetOrderDetail(ctx context.Context, userID, orderID int64) (*model.Order, error) {
	if cached, err := s.cache.GetOrder(ctx, orderID); err == nil {
		return cached, nil
	}

	o, err := s.orders.FindOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.UserID != userID {
		return nil, ErrNotFound
	}

	if err := s.cache.SetOrder(ctx, o); err != nil {
		return nil, fmt.Errorf("cache order: %w", err)
	}
	s.orders.TouchOrder(ctx, orderID)
	return o, nil
}

// GetByID is the internal (gRPC) lookup without an ownership check.
func (s *OrderService) GetOrderByID(ctx context.Context, orderID int64) (*model.Order, error) {
	return s.orders.FindOrder(ctx, orderID)
}

// Create validates and persists a new order.
func (s *OrderService) PlaceOrder(ctx context.Context, userID int64, items []model.OrderItem) (*model.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("empty order")
	}

	var total int64
	for i := range items {
		if items[i].Quantity <= 0 {
			return nil, errors.New("invalid quantity")
		}
		total += items[i].Subtotal()
	}

	o := &model.Order{
		UserID:   userID,
		Status:   model.StatusPending,
		Currency: "CNY",
		Total:    total,
		Items:    items,
	}
	if err := s.orders.InsertOrder(ctx, o); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return o, nil
}

// Cancel cancels a pending order.
func (s *OrderService) CancelOrder(ctx context.Context, userID, orderID int64) error {
	o, err := s.orders.FindOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if o.UserID != userID {
		return ErrNotFound
	}
	if !o.CanTransitionTo(model.StatusCancelled) {
		return ErrInvalidState
	}
	return s.orders.UpdateOrderStatus(ctx, orderID, model.StatusCancelled)
}

// Refund refunds a paid order.
func (s *OrderService) RefundOrder(ctx context.Context, userID, orderID int64) error {
	o, err := s.orders.FindOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if o.UserID != userID {
		return ErrNotFound
	}
	if !o.CanTransitionTo(model.StatusRefunded) {
		return ErrInvalidState
	}

	return s.orders.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Payment{}).
			Where("order_id = ?", orderID).
			Update("status", model.PaymentRefunded).Error; err != nil {
			return err
		}

		if err := s.gateway.Refund(orderID, 0, o.Total); err != nil {
			return fmt.Errorf("gateway refund: %w", err)
		}

		if err := s.ledger.PostDebit(ctx, tx, o.UserID, orderID, o.Total); err != nil {
			return err
		}

		return tx.Model(&model.Order{}).
			Where("id = ?", orderID).
			Update("status", model.StatusRefunded).Error
	})
}

// MarkPaid applies a payment.succeeded event.
func (s *OrderService) ApplyPayment(ctx context.Context, orderID, paymentID int64) error {
	return s.orders.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p model.Payment
		if err := tx.First(&p, paymentID).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Payment{}).
			Where("id = ?", paymentID).
			Update("status", model.PaymentSucceeded).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Order{}).
			Where("id = ?", orderID).
			Update("status", model.StatusPaid).Error; err != nil {
			return err
		}
		return s.ledger.PostCredit(ctx, tx, p.ID, orderID, p.Amount)
	})
}

// BulkUpdateStatus updates the orders selected by the admin console.
func (s *OrderService) BulkUpdateOrderStatus(ctx context.Context, status model.OrderStatus) error {
	return s.orders.BulkSetOrderStatus(ctx, status)
}
