package platform

import "github.com/redis/go-redis/v9"

// OpenRedis creates a Redis client.
func OpenRedis(addr, password string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr, Password: password})
}
