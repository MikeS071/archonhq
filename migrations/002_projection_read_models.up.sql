CREATE OR REPLACE VIEW rm_active_tasks AS
SELECT
  t.tenant_id,
  t.workspace_id,
  t.task_id,
  t.task_family,
  t.title,
  t.status,
  t.created_at
FROM tasks t
WHERE t.status NOT IN ('cancelled', 'completed', 'failed');

CREATE OR REPLACE VIEW rm_approval_queue AS
SELECT
  a.tenant_id,
  a.approval_id,
  a.task_id,
  a.status,
  a.created_at,
  a.decided_at
FROM approval_requests a
WHERE a.status IN ('pending', 'awaiting_decision');

CREATE OR REPLACE VIEW rm_fleet_overview AS
SELECT
  n.tenant_id,
  n.node_id,
  n.operator_id,
  n.runtime_type,
  n.runtime_version,
  n.status,
  n.last_heartbeat_at
FROM nodes n;

CREATE OR REPLACE VIEW rm_node_heartbeat AS
SELECT
  n.tenant_id,
  n.node_id,
  n.last_heartbeat_at,
  n.status
FROM nodes n;

CREATE OR REPLACE VIEW rm_task_trace AS
SELECT
  e.tenant_id,
  e.entity_id AS task_id,
  e.event_type,
  e.occurred_at,
  e.correlation_id,
  e.actor_type,
  e.actor_id
FROM event_records e
WHERE e.entity_type = 'task';

CREATE OR REPLACE VIEW rm_ledger_balances AS
SELECT
  la.tenant_id,
  la.account_id,
  la.owner_type,
  la.owner_id,
  COALESCE(SUM(le.net_amount), 0) AS net_balance,
  COALESCE(SUM(le.reserve_amount), 0) AS reserve_balance
FROM ledger_accounts la
LEFT JOIN ledger_entries le ON la.account_id = le.account_id
GROUP BY la.tenant_id, la.account_id, la.owner_type, la.owner_id;

CREATE OR REPLACE VIEW rm_reliability_summary AS
SELECT DISTINCT ON (tenant_id, subject_type, subject_id)
  tenant_id,
  subject_type,
  subject_id,
  family,
  window_name,
  rf_value,
  created_at
FROM reliability_snapshots
ORDER BY tenant_id, subject_type, subject_id, created_at DESC;

CREATE OR REPLACE VIEW rm_recent_settlements AS
SELECT
  le.tenant_id,
  le.entry_id,
  le.account_id,
  le.result_id,
  le.credited_jw,
  le.net_amount,
  le.status,
  le.created_at
FROM ledger_entries le
WHERE le.event_type = 'ledger.settlement_posted'
ORDER BY le.created_at DESC;
