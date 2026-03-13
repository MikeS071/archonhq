# API_CONTRACTS.md

## Protocol style

- JSON over HTTP
- versioned under `/v1`
- idempotency support for mutating requests
- correlation IDs for tracing
- Clerk bearer auth for humans
- node credentials/tokens for worker nodes

## Core endpoint groups

### /v1/tenants
- POST `/v1/tenants`
- GET `/v1/tenants/{tenant_id}`
- PATCH `/v1/tenants/{tenant_id}`
- GET `/v1/tenants/{tenant_id}/members`

### /v1/workspaces
- POST `/v1/workspaces`
- GET `/v1/workspaces/{workspace_id}`
- GET `/v1/workspaces/{workspace_id}/summary`
- GET `/v1/workspaces/{workspace_id}/tasks`
- GET `/v1/workspaces/{workspace_id}/ledger`

### /v1/nodes
- POST `/v1/nodes/register-intent`
- POST `/v1/nodes/register`
- POST `/v1/nodes/{node_id}/heartbeat`
- GET `/v1/nodes/{node_id}`
- GET `/v1/nodes/{node_id}/leases`

### /v1/tasks
- POST `/v1/tasks`
- GET `/v1/tasks/{task_id}`
- GET `/v1/tasks/feed`
- POST `/v1/tasks/{task_id}/cancel`
- POST `/v1/tasks/{task_id}/decompose`

### /v1/approvals
- GET `/v1/approvals/queue`
- GET `/v1/approvals/{approval_id}`
- POST `/v1/approvals/{approval_id}/approve`
- POST `/v1/approvals/{approval_id}/deny`
- POST `/v1/approvals/{approval_id}/auto-mode`

### /v1/leases
- POST `/v1/leases`
- POST `/v1/leases/{lease_id}/claim`
- POST `/v1/leases/{lease_id}/release`
- POST `/v1/leases/{lease_id}/extend`

### /v1/artifacts
- POST `/v1/artifacts/upload-url`
- POST `/v1/artifacts/register`
- GET `/v1/artifacts/{artifact_id}`
- GET `/v1/artifacts/{artifact_id}/download-url`

### /v1/results
- POST `/v1/results`
- GET `/v1/results/{result_id}`
- GET `/v1/tasks/{task_id}/results`

### /v1/verifications
- POST `/v1/verifications`
- GET `/v1/verifications/{verification_id}`
- GET `/v1/results/{result_id}/verifications`

### /v1/reductions
- POST `/v1/reductions`
- GET `/v1/reductions/{reduction_id}`

### /v1/reliability
- GET `/v1/reliability/subjects/{subject_type}/{subject_id}`
- GET `/v1/operators/{operator_id}/reliability`

### /v1/simulation/scenarios
- POST `/v1/simulation/scenarios`
- GET `/v1/simulation/scenarios`
- GET `/v1/simulation/scenarios/{scenario_id}`
- POST `/v1/simulation/scenarios/{scenario_id}/versions`
- POST `/v1/simulation/scenarios/{scenario_id}/publish`

### /v1/simulation/runs
- POST `/v1/simulation/runs`
- GET `/v1/simulation/runs`
- GET `/v1/simulation/runs/{run_id}`
- POST `/v1/simulation/runs/{run_id}/cancel`
- GET `/v1/simulation/runs/{run_id}/events`
- GET `/v1/simulation/runs/{run_id}/metrics`
- GET `/v1/simulation/runs/{run_id}/findings`
- GET `/v1/simulation/runs/{run_id}/artifacts`

### /v1/simulation/baselines
- POST `/v1/simulation/runs/{run_id}/promote-baseline`
- GET `/v1/simulation/baselines`
- GET `/v1/simulation/baselines/{baseline_id}`
- POST `/v1/simulation/compare`

### /v1/simulation/replays
- POST `/v1/simulation/replays`
- GET `/v1/simulation/replays/{replay_id}`

### /v1/pricing
- POST `/v1/pricing/quote`
- GET `/v1/pricing/rate-cards`
- POST `/v1/pricing/bids`
- GET `/v1/tasks/{task_id}/market`

### /v1/ledger
- GET `/v1/ledger/accounts/{account_id}`
- GET `/v1/ledger/accounts/{account_id}/entries`
- GET `/v1/operators/{operator_id}/earnings-summary`
- GET `/v1/operators/{operator_id}/reserve-holds`

### /v1/policies
- GET `/v1/policies`
- POST `/v1/policies`
- PATCH `/v1/policies/{policy_id}`

### /v1/integrations/paperclip
- POST `/v1/integrations/paperclip/sync`
- GET `/v1/integrations/paperclip/status`

## Example request: create task
```json
{
  "workspace_id": "ws_01",
  "task_family": "research.extract",
  "title": "Extract vendor security claims",
  "description": "Find and structure claims into schema",
  "input_refs": ["art_urls_01"],
  "schema_ref": "schema_extract_vendor_claims_v1",
  "approval_policy": {"mode": "always_required"},
  "execution_policy": {
    "allowed_backends": ["docker"],
    "allowed_toolsets": ["web", "file"],
    "network_policy": "restricted"
  }
}
```

## Example response: create task
```json
{
  "task_id": "task_01",
  "status": "awaiting_approval",
  "approval_request_id": "apr_01"
}
```

## Simulation API rules
- expensive run creation may return `202 Accepted`
- scenario versions are immutable after publish
- simulation artifacts are isolated from production artifact namespaces
- replaying sensitive production traces requires explicit approval and audit logging

## Paperclip integration API rules
- all projection payloads declare `source_of_truth=postgres`
- Paperclip is projection target only and never authoritative for workflow truth
- sync endpoint returns surface counts for workspace, approvals, fleet, reliability, and settlements
- status endpoint reports latest sync event state for tenant-scoped operational visibility
