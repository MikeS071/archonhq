CREATE TABLE IF NOT EXISTS run_telemetry_refs (
  run_telemetry_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  lease_id TEXT NOT NULL REFERENCES leases(lease_id),
  result_id TEXT NOT NULL REFERENCES results(result_id),
  logs_artifact_id TEXT REFERENCES artifacts(artifact_id),
  tool_calls_artifact_id TEXT REFERENCES artifacts(artifact_id),
  metrics_artifact_id TEXT REFERENCES artifacts(artifact_id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_run_telemetry_result_unique
  ON run_telemetry_refs (tenant_id, result_id);

CREATE INDEX IF NOT EXISTS idx_run_telemetry_lease
  ON run_telemetry_refs (tenant_id, lease_id, created_at DESC);
