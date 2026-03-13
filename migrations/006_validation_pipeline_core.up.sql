CREATE TABLE IF NOT EXISTS critic_registry_entries (
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  critic_id TEXT NOT NULL,
  name TEXT NOT NULL,
  stage_name TEXT NOT NULL CHECK (stage_name IN ('plan', 'execution', 'artifact', 'output', 'policy', 'security', 'benchmark', 'reduction')),
  task_family TEXT NOT NULL,
  critic_class TEXT NOT NULL,
  provider_family TEXT NOT NULL,
  failure_mode_class TEXT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
  published_version INT,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, critic_id)
);

CREATE TABLE IF NOT EXISTS validation_runs (
  validation_run_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  task_id TEXT NOT NULL REFERENCES tasks(task_id),
  result_id TEXT,
  validation_tier TEXT NOT NULL CHECK (validation_tier IN ('fast', 'standard', 'high_assurance')),
  status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'needs_review', 'rejected', 'escalated')),
  decision TEXT CHECK (decision IN ('accepted', 'rejected', 'needs_review')),
  acceptance_contract_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS validation_stage_results (
  validation_stage_result_id TEXT PRIMARY KEY,
  validation_run_id TEXT NOT NULL REFERENCES validation_runs(validation_run_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  stage_order INT NOT NULL,
  stage_name TEXT NOT NULL CHECK (stage_name IN ('plan', 'execution', 'artifact', 'output', 'policy', 'security', 'benchmark', 'reduction')),
  critic_id TEXT,
  critic_class TEXT,
  decision TEXT NOT NULL CHECK (decision IN ('accepted', 'rejected', 'needs_review')),
  score DOUBLE PRECISION,
  evidence_refs_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  retry_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (validation_run_id, stage_order, critic_id, retry_count)
);

CREATE TABLE IF NOT EXISTS validation_escalations (
  validation_escalation_id TEXT PRIMARY KEY,
  validation_run_id TEXT NOT NULL REFERENCES validation_runs(validation_run_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  reason TEXT NOT NULL,
  escalated_by TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved', 'cancelled')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  resolved_at TIMESTAMPTZ
);
