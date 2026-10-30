package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/topolink/topolink-example/internal/model"
)

// Cache stores hot order snapshots and idempotency locks.
type Cache struct {
	rdb *redis.Client
}

// NewCache wraps a Redis client.
func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) orderKey(id int64) string {
	return fmt.Sprintf("order:%d", id)
}

// GetOrder returns a cached order, if present.
func (c *Cache) GetOrder(ctx context.Context, id int64) (*model.Order, error) {
	data, err := c.rdb.Get(ctx, c.orderKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var o model.Order
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

// SetOrder caches an order snapshot.
func (c *Cache) SetOrder(ctx context.Context, o *model.Order) error {
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, c.orderKey(o.ID), data, 0).Err()
}

// AcquireLock takes a short-lived distributed lock.
func (c *Cache) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, key, "1", ttl).Result()
}
