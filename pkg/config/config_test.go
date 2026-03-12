package config

import "testing"

func TestLoadSuccessAndGuards(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://archon:dev@localhost:5432/archonhq?sslmode=disable")
	t.Setenv("NATS_URL", "nats://localhost:4222")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("WORKER_RUNTIME", "hermes")
	t.Setenv("SETTLEMENT_MODE", "ledger_only")
	t.Setenv("PORT", "8081")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Port != "8081" || cfg.WorkerRuntime != "hermes" || cfg.SettlementMode != "ledger_only" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadValidationFailures(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatalf("expected missing database url error")
	}

	t.Setenv("DATABASE_URL", "postgres://archon:dev@localhost:5432/archonhq?sslmode=disable")
	t.Setenv("WORKER_RUNTIME", "not-hermes")
	if _, err := Load(); err == nil {
		t.Fatalf("expected runtime guard error")
	}

	t.Setenv("WORKER_RUNTIME", "hermes")
	t.Setenv("SETTLEMENT_MODE", "external")
	if _, err := Load(); err == nil {
		t.Fatalf("expected settlement guard error")
	}
}
