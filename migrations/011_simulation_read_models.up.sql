CREATE TABLE IF NOT EXISTS rm_simulation_run_summary (
  tenant_id TEXT NOT NULL,
  run_id TEXT PRIMARY KEY,
  scenario_id TEXT NOT NULL,
  scenario_version INT NOT NULL,
  run_mode TEXT NOT NULL,
  status TEXT NOT NULL,
  finding_count INT NOT NULL DEFAULT 0,
  critical_finding_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rm_simulation_findings (
  tenant_id TEXT NOT NULL,
  simulation_finding_id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  scenario_id TEXT NOT NULL,
  finding_type TEXT NOT NULL,
  severity TEXT NOT NULL,
  summary TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS rm_simulation_baseline_diffs (
  tenant_id TEXT NOT NULL,
  compare_id TEXT PRIMARY KEY,
  baseline_id TEXT NOT NULL,
  candidate_run_id TEXT NOT NULL,
  verdict TEXT NOT NULL,
  diff_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rm_simulation_risk_heatmap (
  tenant_id TEXT NOT NULL,
  scenario_id TEXT NOT NULL,
  risk_bucket TEXT NOT NULL,
  score DOUBLE PRECISION NOT NULL,
  sample_size INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, scenario_id, risk_bucket)
);
