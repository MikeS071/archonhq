CREATE TABLE IF NOT EXISTS simulation_run_steps (
  simulation_run_step_id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES simulation_runs(run_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  step_index INT NOT NULL,
  step_type TEXT NOT NULL,
  status TEXT NOT NULL,
  payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (run_id, step_index)
);

CREATE TABLE IF NOT EXISTS simulation_run_artifacts (
  simulation_run_artifact_id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES simulation_runs(run_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  artifact_kind TEXT NOT NULL,
  artifact_ref TEXT NOT NULL,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS simulation_run_metrics (
  simulation_run_metric_id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES simulation_runs(run_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  metric_name TEXT NOT NULL,
  metric_value DOUBLE PRECISION NOT NULL,
  unit TEXT,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS simulation_findings (
  simulation_finding_id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES simulation_runs(run_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  finding_type TEXT NOT NULL,
  severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
  summary TEXT NOT NULL,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS simulation_baselines (
  baseline_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  scenario_id TEXT NOT NULL,
  scenario_version INT NOT NULL,
  run_id TEXT NOT NULL REFERENCES simulation_runs(run_id),
  reason TEXT,
  promoted_by TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS simulation_policy_snapshots (
  simulation_policy_snapshot_id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES simulation_runs(run_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  policy_family TEXT NOT NULL,
  policy_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
