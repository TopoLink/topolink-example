// Package event contains Kafka consumers.
package event

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"

	"github.com/topolink/topolink-example/internal/service"
)

// PaymentSucceeded is the payment.succeeded event payload.
type PaymentSucceeded struct {
	OrderID   int64 `json:"order_id"`
	PaymentID int64 `json:"payment_id"`
}

// PaymentConsumer consumes payment.succeeded events and marks orders paid.
type PaymentConsumer struct {
	reader *kafka.Reader
	orders *service.OrderService
	log    *slog.Logger
}

// NewPaymentConsumer builds a consumer for the payment topic.
func NewPaymentConsumer(
	brokers []string,
	topic, group string,
	orders *service.OrderService,
	logger *slog.Logger,
) *PaymentConsumer {
	return &PaymentConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: group,
		}),
		orders: orders,
		log:    logger,
	}
}

// Close releases the Kafka reader.
func (c *PaymentConsumer) Close() error {
	return c.reader.Close()
}

// Consume processes events until the context is canceled.
func (c *PaymentConsumer) Consume(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var evt PaymentSucceeded
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			c.log.Error("invalid payment event", "err", err)
			continue
		}

		if err := c.orders.ApplyPayment(ctx, evt.OrderID, evt.PaymentID); err != nil {
			c.log.Error("mark paid failed", "order_id", evt.OrderID, "err", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Error("commit failed", "err", err)
		}
	}
}
