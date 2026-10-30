// Command api runs the HTTP and gRPC servers.
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/topolink/topolink-example/internal/config"
	"github.com/topolink/topolink-example/internal/gateway"
	"github.com/topolink/topolink-example/internal/platform"
	"github.com/topolink/topolink-example/internal/repository"
	"github.com/topolink/topolink-example/internal/service"
	grpcapi "github.com/topolink/topolink-example/internal/transport/grpcapi"
	httpapi "github.com/topolink/topolink-example/internal/transport/httpapi"
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
	settlementSvc := service.NewSettlementService(ledger, payments)

	router := httpapi.NewRouter(orderSvc, settlementSvc, cfg.JWTSecret, logger)
	httpSrv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	grpcSrv := grpc.NewServer()
	grpcapi.RegisterOrderServiceServer(grpcSrv, &grpcapi.OrderServer{Orders: orderSvc})

	go func() {
		logger.Info("http server listening", "addr", cfg.HTTPAddr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server", "err", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", cfg.GRPCAddr)
		if err != nil {
			logger.Error("grpc listen", "err", err)
			return
		}
		logger.Info("grpc server listening", "addr", cfg.GRPCAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Error("grpc server", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Error("http shutdown", "err", err)
	}
	grpcSrv.GracefulStop()
	logger.Info("stopped")
}
