package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/topolink/topolink-example/internal/model"
	"github.com/topolink/topolink-example/internal/service"
)

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (a *API) listCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []string{"sku-1", "sku-2"}})
}

func (a *API) listOrders(c *gin.Context) {
	userID := c.GetInt64("user_id")
	limit := clamp(c.DefaultQuery("limit", "20"), 20)
	offset := clamp(c.DefaultQuery("offset", "0"), 0)

	orders, err := a.orders.ListOrders(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list orders failed"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (a *API) getOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	o, err := a.orders.GetOrderDetail(c.Request.Context(), userID, orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (a *API) createOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		Items []model.OrderItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	o, err := a.orders.PlaceOrder(c.Request.Context(), userID, req.Items)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

func (a *API) cancelOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := a.orders.CancelOrder(c.Request.Context(), userID, orderID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) refundOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := a.orders.RefundOrder(c.Request.Context(), userID, orderID); err != nil {
		if err == service.ErrInvalidState {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) adminListOrders(c *gin.Context) {
	orders, err := a.orders.ListAllOrders(c.Request.Context(), 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list orders failed"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (a *API) bulkStatus(c *gin.Context) {
	var req struct {
		Status model.OrderStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.orders.BulkUpdateOrderStatus(c.Request.Context(), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bulk update failed"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) runSettlement(c *gin.Context) {
	var req struct {
		BatchID string `json:"batch_id"`
		Amount  int64  `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.settlements.RunSettlement(c.Request.Context(), req.BatchID, req.Amount); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) legacyListOrders(c *gin.Context) {
	orders, err := a.orders.ListAllOrders(c.Request.Context(), 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list orders failed"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func clamp(raw string, fallback int) int {
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return fallback
	}
	if v > 100 {
		return 100
	}
	return v
}
