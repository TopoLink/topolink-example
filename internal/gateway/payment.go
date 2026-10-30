// Package gateway talks to the external payment provider.
package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// PaymentGateway is the subset of the provider API used by the service.
type PaymentGateway interface {
	Refund(orderID, paymentID, amount int64) error
	Settle(batchID string, amount int64) error
}

// PaymentClient is the HTTP implementation.
type PaymentClient struct {
	baseURL string
}

// NewPaymentClient builds a client for the provider base URL.
func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{baseURL: baseURL}
}

// Refund asks the provider to refund a payment.
func (c *PaymentClient) Refund(orderID, paymentID, amount int64) error {
	body, err := json.Marshal(map[string]any{
		"order_id":   orderID,
		"payment_id": paymentID,
		"amount":     amount,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(c.baseURL+"/v1/refunds", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("refund request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("refund rejected: %s", resp.Status)
	}
	return nil
}

// Settle pushes a settlement batch to the provider.
func (c *PaymentClient) Settle(batchID string, amount int64) error {
	body, err := json.Marshal(map[string]any{
		"batch_id": batchID,
		"amount":   amount,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(c.baseURL+"/v1/settlements", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("settlement request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("settlement rejected: %s", resp.Status)
	}
	return nil
}
