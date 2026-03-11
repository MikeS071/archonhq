-- 000_full_schema.sql

CREATE TABLE tenants (
  tenant_id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  signup_mode TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memberships (
  membership_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  user_id TEXT NOT NULL,
  role TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspaces (
  workspace_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE operators (
  operator_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  user_id TEXT,
  display_name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active'
);

CREATE TABLE nodes (
  node_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  operator_id TEXT NOT NULL REFERENCES operators(operator_id),
  public_key TEXT NOT NULL,
  runtime_type TEXT NOT NULL,
  runtime_version TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  last_heartbeat_at TIMESTAMPTZ
);

CREATE TABLE agents (
  agent_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  node_id TEXT NOT NULL REFERENCES nodes(node_id),
  runtime_name TEXT NOT NULL,
  lineage_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  status TEXT NOT NULL DEFAULT 'active'
);

CREATE TABLE provider_credentials (
  credential_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  operator_id TEXT NOT NULL REFERENCES operators(operator_id),
  provider_name TEXT NOT NULL,
  secret_ref TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active'
);

CREATE TABLE policy_bundles (
  policy_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  workspace_id TEXT,
  family TEXT,
  version INT NOT NULL DEFAULT 1,
  policy_json JSONB NOT NULL
);

CREATE TABLE tasks (
  task_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  workspace_id TEXT NOT NULL REFERENCES workspaces(workspace_id),
  task_family TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL,
  schema_ref TEXT,
  merge_strategy TEXT,
  pricing_mode TEXT,
  approval_mode TEXT,
  created_by TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE leases (
  lease_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  task_id TEXT NOT NULL REFERENCES tasks(task_id),
  node_id TEXT NOT NULL REFERENCES nodes(node_id),
  agent_id TEXT,
  attempt INT NOT NULL DEFAULT 1,
  status TEXT NOT NULL,
  approval_state TEXT NOT NULL,
  granted_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ,
  execution_policy_snapshot_json JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE approval_requests (
  approval_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  task_id TEXT NOT NULL REFERENCES tasks(task_id),
  lease_id TEXT,
  status TEXT NOT NULL,
  requested_by TEXT,
  decided_by TEXT,
  decision_reason TEXT,
  payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  decided_at TIMESTAMPTZ
);

CREATE TABLE artifacts (
  artifact_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  workspace_id TEXT NOT NULL REFERENCES workspaces(workspace_id),
  blob_ref TEXT NOT NULL,
  sha256 TEXT NOT NULL,
  media_type TEXT NOT NULL,
  size_bytes BIGINT NOT NULL,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE results (
  result_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  task_id TEXT NOT NULL REFERENCES tasks(task_id),
  lease_id TEXT NOT NULL REFERENCES leases(lease_id),
  node_id TEXT NOT NULL REFERENCES nodes(node_id),
  agent_id TEXT,
  status TEXT NOT NULL,
  metering_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  quality_inputs_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  signature TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE result_output_refs (
  result_id TEXT NOT NULL REFERENCES results(result_id),
  artifact_id TEXT NOT NULL REFERENCES artifacts(artifact_id),
  PRIMARY KEY (result_id, artifact_id)
);

CREATE TABLE verifications (
  verification_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  result_id TEXT NOT NULL REFERENCES results(result_id),
  verifier_type TEXT NOT NULL,
  verifier_id TEXT,
  score DOUBLE PRECISION,
  decision TEXT NOT NULL,
  report_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE reductions (
  reduction_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  task_id TEXT NOT NULL REFERENCES tasks(task_id),
  strategy TEXT NOT NULL,
  output_state_ref TEXT,
  decision_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE reliability_snapshots (
  snapshot_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  subject_type TEXT NOT NULL,
  subject_id TEXT NOT NULL,
  family TEXT,
  window_name TEXT NOT NULL,
  rf_value DOUBLE PRECISION NOT NULL,
  components_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE price_quotes (
  quote_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  task_id TEXT NOT NULL REFERENCES tasks(task_id),
  strategy_name TEXT NOT NULL,
  quote_json JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE rate_snapshots (
  rate_snapshot_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  task_id TEXT NOT NULL REFERENCES tasks(task_id),
  result_id TEXT,
  strategy_name TEXT NOT NULL,
  rate_value NUMERIC(18,8) NOT NULL,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ledger_accounts (
  account_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  owner_type TEXT NOT NULL,
  owner_id TEXT NOT NULL,
  currency TEXT NOT NULL DEFAULT 'JWUSD',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ledger_entries (
  entry_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  account_id TEXT NOT NULL REFERENCES ledger_accounts(account_id),
  event_type TEXT NOT NULL,
  result_id TEXT,
  raw_jw NUMERIC(18,8),
  credited_jw NUMERIC(18,8),
  rate NUMERIC(18,8),
  gross_amount NUMERIC(18,8),
  reserve_amount NUMERIC(18,8),
  net_amount NUMERIC(18,8),
  status TEXT NOT NULL,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE reserve_holds (
  reserve_hold_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  ledger_entry_id TEXT NOT NULL REFERENCES ledger_entries(entry_id),
  status TEXT NOT NULL,
  release_after TIMESTAMPTZ NOT NULL,
  released_at TIMESTAMPTZ
);

CREATE TABLE event_records (
  event_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  workspace_id TEXT,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  event_version INT NOT NULL,
  actor_type TEXT,
  actor_id TEXT,
  correlation_id TEXT,
  idempotency_key TEXT,
  payload_json JSONB NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_memberships_tenant ON memberships(tenant_id, created_at DESC);
CREATE INDEX idx_workspaces_tenant ON workspaces(tenant_id, created_at DESC);
CREATE INDEX idx_nodes_tenant ON nodes(tenant_id, created_at DESC);
CREATE INDEX idx_tasks_tenant_workspace ON tasks(tenant_id, workspace_id, created_at DESC);
CREATE INDEX idx_leases_task ON leases(task_id, granted_at DESC);
CREATE INDEX idx_approvals_task ON approval_requests(task_id, created_at DESC);
CREATE INDEX idx_artifacts_workspace ON artifacts(workspace_id, created_at DESC);
CREATE INDEX idx_results_task ON results(task_id, created_at DESC);
CREATE INDEX idx_verifications_result ON verifications(result_id, created_at DESC);
CREATE INDEX idx_reliability_subject ON reliability_snapshots(tenant_id, subject_type, subject_id, created_at DESC);
CREATE INDEX idx_ledger_account_created ON ledger_entries(account_id, created_at DESC);
CREATE INDEX idx_events_tenant_created ON event_records(tenant_id, created_at DESC);
CREATE INDEX idx_events_entity_created ON event_records(entity_type, entity_id, created_at DESC);
