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
- simulation_scenarios
- simulation_scenario_versions
- simulation_runs
- simulation_run_steps
- simulation_run_actors
- simulation_run_artifacts
- simulation_run_metrics
- simulation_findings
- simulation_baselines
- simulation_policy_snapshots
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
- rm_simulation_run_summary
- rm_simulation_findings
- rm_simulation_baseline_diffs
- rm_simulation_risk_heatmap
