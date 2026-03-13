CREATE TABLE IF NOT EXISTS simulation_scenarios (
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  scenario_id TEXT NOT NULL,
  scope TEXT NOT NULL CHECK (scope IN ('platform', 'tenant')),
  name TEXT NOT NULL,
  goal TEXT,
  status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
  published_version INT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, scenario_id)
);

CREATE TABLE IF NOT EXISTS simulation_scenario_versions (
  tenant_id TEXT NOT NULL,
  scenario_id TEXT NOT NULL,
  version INT NOT NULL,
  spec_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, scenario_id, version),
  FOREIGN KEY (tenant_id, scenario_id) REFERENCES simulation_scenarios(tenant_id, scenario_id)
);

CREATE TABLE IF NOT EXISTS simulation_runs (
  run_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  scenario_id TEXT NOT NULL,
  scenario_version INT NOT NULL,
  run_mode TEXT NOT NULL CHECK (run_mode IN ('deterministic_stub', 'sampled_synthetic', 'runtime_backed')),
  status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'cancelled', 'failed')),
  seed TEXT,
  scale_profile TEXT,
  budget_limit_jw NUMERIC(18,8),
  timebox_seconds INT,
  policy_overrides_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  FOREIGN KEY (tenant_id, scenario_id, scenario_version)
    REFERENCES simulation_scenario_versions(tenant_id, scenario_id, version)
);

CREATE TABLE IF NOT EXISTS simulation_run_actors (
  run_actor_id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES simulation_runs(run_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  actor_type TEXT NOT NULL,
  actor_id TEXT,
  profile_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
