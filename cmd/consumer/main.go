// Command consumer runs the payment.succeeded Kafka consumer.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/topolink/topolink-example/internal/config"
	"github.com/topolink/topolink-example/internal/event"
	"github.com/topolink/topolink-example/internal/gateway"
	"github.com/topolink/topolink-example/internal/platform"
	"github.com/topolink/topolink-example/internal/repository"
	"github.com/topolink/topolink-example/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	logger := platform.NewLogger(cfg.Env)

	db, err := platform.OpenDB(cfg.DatabaseURL, cfg.Env)
	if err != nil {
		logger.Error("database", "err", err)
		os.Exit(1)
	}
	rdb := platform.OpenRedis(cfg.RedisAddr, cfg.RedisPassword)

	orders := repository.NewOrderRepository(db)
	ledger := repository.NewLedgerRepository(db)
	cache := repository.NewCache(rdb)
	payments := gateway.NewPaymentClient(cfg.PaymentBaseURL)
	orderSvc := service.NewOrderService(orders, ledger, cache, payments)

	consumer := event.NewPaymentConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
		orderSvc,
		logger,
	)
	defer consumer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := consumer.Consume(ctx); err != nil && ctx.Err() == nil {
		logger.Error("consumer", "err", err)
		os.Exit(1)
	}
}
