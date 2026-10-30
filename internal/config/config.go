// Package config loads runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config is the service configuration.
type Config struct {
	Env             string
	HTTPAddr        string
	GRPCAddr        string
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
	KafkaBrokers    []string
	KafkaTopic      string
	KafkaGroupID    string
	JWTSecret       string
	PaymentBaseURL  string
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment and validates required values.
func Load() (Config, error) {
	cfg := Config{
		Env:             env("APP_ENV", "development"),
		HTTPAddr:        env("HTTP_ADDR", ":8080"),
		GRPCAddr:        env("GRPC_ADDR", ":9090"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisAddr:       env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
		KafkaBrokers:    split(env("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:      env("KAFKA_TOPIC", "payment.succeeded"),
		KafkaGroupID:    env("KAFKA_GROUP_ID", "order-service"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		PaymentBaseURL:  env("PAYMENT_BASE_URL", "http://localhost:7070"),
		ShutdownTimeout: duration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func split(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func duration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
