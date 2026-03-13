CREATE INDEX IF NOT EXISTS idx_simulation_scenarios_lookup
  ON simulation_scenarios (tenant_id, scenario_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_simulation_runs_lookup
  ON simulation_runs (tenant_id, scenario_id, scenario_version, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_simulation_run_metrics_lookup
  ON simulation_run_metrics (tenant_id, run_id, metric_name, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_simulation_findings_lookup
  ON simulation_findings (tenant_id, run_id, finding_type, severity, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_simulation_baselines_lookup
  ON simulation_baselines (tenant_id, scenario_id, scenario_version, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_simulation_runs_active
  ON simulation_runs (tenant_id, created_at DESC)
  WHERE status IN ('queued', 'running');
