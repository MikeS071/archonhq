CREATE INDEX IF NOT EXISTS idx_validation_runs_lookup
  ON validation_runs (tenant_id, task_id, result_id, validation_tier, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_validation_stage_results_lookup
  ON validation_stage_results (validation_run_id, stage_name, decision, critic_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_critic_registry_lookup
  ON critic_registry_entries (tenant_id, critic_class, task_family, enabled, created_at DESC);

CREATE TABLE IF NOT EXISTS rm_validation_run_summary (
  tenant_id TEXT NOT NULL,
  validation_run_id TEXT PRIMARY KEY,
  task_id TEXT NOT NULL,
  result_id TEXT,
  validation_tier TEXT NOT NULL,
  status TEXT NOT NULL,
  decision TEXT,
  stage_count INT NOT NULL DEFAULT 0,
  needs_review_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rm_validation_escalation_queue (
  tenant_id TEXT NOT NULL,
  validation_escalation_id TEXT PRIMARY KEY,
  validation_run_id TEXT NOT NULL,
  task_id TEXT,
  reason TEXT NOT NULL,
  status TEXT NOT NULL,
  escalated_by TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  resolved_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS rm_critic_effectiveness (
  tenant_id TEXT NOT NULL,
  critic_id TEXT NOT NULL,
  task_family TEXT NOT NULL,
  stage_name TEXT NOT NULL,
  accepted_count BIGINT NOT NULL DEFAULT 0,
  rejected_count BIGINT NOT NULL DEFAULT 0,
  needs_review_count BIGINT NOT NULL DEFAULT 0,
  disagreement_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, critic_id, task_family, stage_name)
);
