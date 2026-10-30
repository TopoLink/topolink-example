package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/topolink/topolink-example/internal/gateway"
	"github.com/topolink/topolink-example/internal/repository"
)

// SettlementService runs daily settlement batches.
type SettlementService struct {
	mu      sync.Mutex
	ledger  *repository.LedgerRepository
	gateway gateway.PaymentGateway
}

// NewSettlementService wires the settlement service.
func NewSettlementService(ledger *repository.LedgerRepository, gateway gateway.PaymentGateway) *SettlementService {
	return &SettlementService{ledger: ledger, gateway: gateway}
}

// Run settles one batch. The mutex prevents concurrent runs from overlapping.
func (s *SettlementService) RunSettlement(ctx context.Context, batchID string, amount int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.gateway.Settle(batchID, amount); err != nil {
		return fmt.Errorf("settle batch %s: %w", batchID, err)
	}
	return s.ledger.MarkSettlementSettled(ctx, batchID, amount)
}
