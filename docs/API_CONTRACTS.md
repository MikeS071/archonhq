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

### /v1/acceptance-contract-templates
- POST `/v1/acceptance-contract-templates`
- GET `/v1/acceptance-contract-templates`
- GET `/v1/acceptance-contract-templates/{template_id}`
- POST `/v1/acceptance-contract-templates/{template_id}/versions`
- POST `/v1/acceptance-contract-templates/{template_id}/publish`

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

### /v1/critics
- GET `/v1/critics`
- GET `/v1/critics/{critic_id}`
- POST `/v1/critics`
- POST `/v1/critics/{critic_id}/versions`
- POST `/v1/critics/{critic_id}/publish`

### /v1/validation-runs
- POST `/v1/tasks/{task_id}/validation-runs`
- GET `/v1/tasks/{task_id}/validation-runs`
- GET `/v1/validation/dashboard`
- GET `/v1/validation-runs/{validation_run_id}`
- GET `/v1/validation-runs/{validation_run_id}/stages`
- POST `/v1/validation-runs/{validation_run_id}/escalate`

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
- GET `/v1/simulation/dashboard`

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

### /v1/market/profiles
- POST `/v1/market/profiles`
- GET `/v1/market/profiles/{profile_id}`
- PATCH `/v1/market/profiles/{profile_id}`
- GET `/v1/market/profiles/{profile_id}/reputation`

### /v1/market/listings
- POST `/v1/market/listings`
- GET `/v1/market/listings`
- GET `/v1/market/listings/{listing_id}`
- POST `/v1/market/listings/{listing_id}/publish`
- POST `/v1/market/listings/{listing_id}/cancel`

### /v1/market/claims
- POST `/v1/market/listings/{listing_id}/claims`
- POST `/v1/market/claims/{claim_id}/withdraw`
- POST `/v1/market/claims/{claim_id}/award`

### /v1/market/bids
- POST `/v1/market/listings/{listing_id}/bids`
- POST `/v1/market/bids/{bid_id}/accept`

### /v1/market/funding-accounts
- POST `/v1/market/funding-accounts`
- GET `/v1/market/funding-accounts/{account_id}`

### /v1/market/escrows
- GET `/v1/market/escrows/{escrow_id}`
- POST `/v1/market/escrows/{escrow_id}/release`
- POST `/v1/market/escrows/{escrow_id}/refund`

### /v1/payout-accounts
- POST `/v1/payout-accounts`
- GET `/v1/payout-accounts/{payout_account_id}`

### /v1/payouts
- POST `/v1/payouts`
- GET `/v1/payouts/{payout_id}`

### /v1/market/disputes
- POST `/v1/market/disputes`
- GET `/v1/market/disputes/{dispute_id}`
- POST `/v1/market/disputes/{dispute_id}/resolve`
- POST `/v1/market/disputes/{dispute_id}/appeal`

## Example request: create task
```json
{
  "workspace_id": "ws_01",
  "task_family": "research.extract",
  "title": "Extract vendor security claims",
  "description": "Find and structure claims into schema",
  "input_refs": ["art_urls_01"],
  "schema_ref": "schema_extract_vendor_claims_v1",
  "acceptance_contract_template_id": "act_research_extract_standard_v1",
  "validation_tier": "standard",
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
  "approval_request_id": "apr_01",
  "validation_tier": "standard",
  "acceptance_contract": {
    "contract_id": "ac_01",
    "contract_version": 1,
    "required_critic_classes": [
      "plan_soundness",
      "evidence_completeness",
      "output_correctness"
    ]
  }
}
```

## Acceptance contract API rules
- task families marked as trust-sensitive require either an inline acceptance contract or a published template reference
- template versions are immutable after publish
- contract overrides must be explicit, bounded, and audit logged
- accepted tasks snapshot the final contract payload used for execution and validation

## Validation API rules
- validation runs are append-only decision records for a given task or result
- each validation run snapshots the selected contract, tier, and critic bundle
- stage results must record critic identity, decision, score, and evidence references
- producer completion signals are insufficient for accepted-state transitions when a validation run is required

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

## M7 advanced workload API rules
- decompose endpoint must return merge strategy, planned child shards, and simulation entrypoints
- approval auto-mode must persist bounded loop guardrails (`max_iterations`, `budget_limit_jw`, approval gate flags)
- verification creation must emit auditable hook outputs and lineage metadata
- reduction creation must enforce supported merge strategies and provide lineage + simulation entrypoints
- task market endpoint must expose verification/reduction signals for operator decisioning

## Judgment-layer API rules
- `validation_tier` accepts `fast`, `standard`, or `high_assurance`
- `high_assurance` workloads may require cross-provider or cross-failure-mode critic diversity by policy
- validation stage failures may return `409 Conflict` when work cannot advance without retry or escalation
- validation escalation endpoints must produce auditable operator-visible state
- validation dashboards must expose effectiveness and escalation-queue visibility for operators

## Simulation operator API rules
- simulation dashboards must expose run/finding/baseline/risk summaries for tenant-scoped operator views

## Open-market API rules
- market listings require funded reserve checks before publish
- work class must be declared as `public_open`, `public_sealed`, `restricted_market`, or `private_tenant_only`
- sealed work must not expose protected artifacts prior to award and policy checks
- payout endpoints are post-v1 and must not mutate accepted task history directly
- dispute resolution may alter escrow outcomes and market reputation, but not erase validation lineage
