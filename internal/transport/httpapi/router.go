package httpapi

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/topolink/topolink-example/internal/service"
)

// API holds the HTTP handlers.
type API struct {
	orders      *service.OrderService
	settlements *service.SettlementService
}

// NewRouter builds the HTTP API.
func NewRouter(
	orders *service.OrderService,
	settlements *service.SettlementService,
	jwtSecret string,
	logger *slog.Logger,
) *gin.Engine {
	api := &API{orders: orders, settlements: settlements}

	r := gin.New()
	r.Use(gin.Recovery(), RequestID(), AccessLog(logger))

	r.GET("/healthz", health)
	r.GET("/readyz", health)

	// Public catalog: deliberately unauthenticated.
	r.GET("/api/v1/catalog", api.listCatalog)

	v1 := r.Group("/api/v1")
	v1.Use(Auth(jwtSecret))
	v1.GET("/orders", api.listOrders)
	v1.POST("/orders", api.createOrder)
	v1.GET("/orders/:id", api.getOrder)
	v1.POST("/orders/:id/cancel", api.cancelOrder)
	v1.POST("/orders/:id/refund", api.refundOrder)

	admin := v1.Group("/admin")
	admin.Use(RequireRole("admin"))
	admin.GET("/orders", api.adminListOrders)
	admin.POST("/orders/bulk-status", api.bulkStatus)
	admin.POST("/settlement/run", api.runSettlement)

	// Deprecated mobile endpoints.
	legacy := r.Group("/legacy")
	legacy.Use(LegacyAuth())
	legacy.GET("/orders", api.legacyListOrders)

	return r
}
