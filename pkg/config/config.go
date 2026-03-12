package config

import (
	"fmt"
	"os"
)

// Config centralizes process-level configuration.
type Config struct {
	AppEnv              string
	Port                string
	DatabaseURL         string
	NATSURL             string
	RedisURL            string
	WorkerRuntime       string
	SettlementMode      string
	S3Endpoint          string
	S3AccessKey         string
	S3SecretKey         string
	S3Bucket            string
	S3Region            string
	ClerkSecretKey      string
	ClerkPublishableKey string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		NATSURL:             os.Getenv("NATS_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		WorkerRuntime:       getEnv("WORKER_RUNTIME", "hermes"),
		SettlementMode:      getEnv("SETTLEMENT_MODE", "ledger_only"),
		S3Endpoint:          os.Getenv("S3_ENDPOINT"),
		S3AccessKey:         os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:         os.Getenv("S3_SECRET_KEY"),
		S3Bucket:            os.Getenv("S3_BUCKET"),
		S3Region:            os.Getenv("S3_REGION"),
		ClerkSecretKey:      os.Getenv("CLERK_SECRET_KEY"),
		ClerkPublishableKey: os.Getenv("CLERK_PUBLISHABLE_KEY"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.NATSURL == "" {
		cfg.NATSURL = "nats://localhost:4222"
	}
	if cfg.RedisURL == "" {
		cfg.RedisURL = "redis://localhost:6379"
	}
	if cfg.WorkerRuntime != "hermes" {
		return Config{}, fmt.Errorf("WORKER_RUNTIME must be 'hermes' in v1")
	}
	if cfg.SettlementMode != "ledger_only" {
		return Config{}, fmt.Errorf("SETTLEMENT_MODE must be 'ledger_only' in v1")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
