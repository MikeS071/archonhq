# DB_SCHEMA_DRAFT.md

## Principles
- include `tenant_id` on all tenant-scoped records
- store large bytes outside Postgres
- use append-only event table
- build read models for dashboards
- keep operational tables normalized

## Core tables
- tenants
- memberships
- workspaces
- operators
- nodes
- agents
- provider_credentials
- policy_bundles
- tasks
- leases
- approval_requests
- artifacts
- results
- result_output_refs
- verifications
- reductions
- reliability_snapshots
- price_quotes
- rate_snapshots
- ledger_accounts
- ledger_entries
- reserve_holds
- event_records

## Recommended read models
- rm_active_tasks
- rm_approval_queue
- rm_fleet_overview
- rm_node_heartbeat
- rm_task_trace
- rm_ledger_balances
- rm_reliability_summary
- rm_recent_settlements
