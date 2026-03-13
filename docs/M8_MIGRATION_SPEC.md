# M8 Migration Spec (Simulation and Assurance)

## Purpose

Define the migration units needed to introduce simulation storage, read models, and isolation controls without coupling schema DDL and large data backfills.

This is a spec for migration implementation. It does not execute DDL by itself.

## Migration Units

### 005 `simulation_schema_core`

Files:
- `migrations/005_simulation_schema_core.up.sql`
- `migrations/005_simulation_schema_core.down.sql`

Creates:
- `simulation_scenarios`
- `simulation_scenario_versions`
- `simulation_runs`
- `simulation_run_actors`

Core constraints:
- `scope` check (`platform`, `tenant`)
- `status` checks for scenario and run lifecycle
- `run_mode` check (`deterministic_stub`, `sampled_synthetic`, `runtime_backed`)
- unique `(scenario_id, version)` for immutable versioning

### 006 `simulation_schema_runtime`

Files:
- `migrations/006_simulation_schema_runtime.up.sql`
- `migrations/006_simulation_schema_runtime.down.sql`

Creates:
- `simulation_run_steps`
- `simulation_run_artifacts`
- `simulation_run_metrics`
- `simulation_findings`
- `simulation_baselines`
- `simulation_policy_snapshots`

Core constraints:
- ordered uniqueness `(run_id, step_index)`
- severity check (`low`, `medium`, `high`, `critical`)
- foreign keys from run-scoped records to `simulation_runs`

### 007 `simulation_indexes_and_isolation`

Files:
- `migrations/007_simulation_indexes_and_isolation.up.sql`
- `migrations/007_simulation_indexes_and_isolation.down.sql`

Adds:
- lookup indexes on `scenario_id`, `scenario_version_id`, `run_id`, `tenant_id`, `status`, `created_at`
- metrics/finding indexes by `run_id`, `metric_name`, `finding_type`, `severity`
- baseline indexes by `scenario_version_id`, `created_at`
- optional partial indexes for active/running runs

Isolation checks:
- tenant-scoped access predicates supported by compound indexes
- explicit separation from production table foreign-key graph

### 008 `simulation_read_models`

Files:
- `migrations/008_simulation_read_models.up.sql`
- `migrations/008_simulation_read_models.down.sql`

Creates read models:
- `rm_simulation_run_summary`
- `rm_simulation_findings`
- `rm_simulation_baseline_diffs`
- `rm_simulation_risk_heatmap`

Notes:
- materializers populate these tables from `simulation.*` events
- no direct writes from runtime handlers

## Backfill/Post-Deploy Work

Handled separately from schema migrations:
- seed v1 scenario library records
- optional baseline seed records for deterministic CI scenarios
- projection backfill jobs

These are data migrations and should be implemented as dedicated scripts/jobs after 005-008 are live.

## Rollback Posture

- Production posture remains forward-fix.
- `.down.sql` files exist for local/dev recovery.
- High-volume indexes should use non-blocking strategies where applicable.

## Verification Checklist

- `go test ./...` still passes after schema changes and repository wiring.
- migration apply/revert smoke in local database.
- simulation tables do not appear in production read-model materializers by default.
- simulation artifacts and events remain namespace-isolated.
