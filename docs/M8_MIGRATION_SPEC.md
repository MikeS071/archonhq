# M8 Migration Spec (Judgment, Simulation, and Assurance)

## Purpose

Define the migration units needed to introduce judgment-layer storage, simulation storage, read models, and isolation controls without coupling schema DDL and large data backfills.

This is a spec for migration implementation. It does not execute DDL by itself.

## Migration Units

### 005 `acceptance_contracts_core`

Files:
- `migrations/005_acceptance_contracts_core.up.sql`
- `migrations/005_acceptance_contracts_core.down.sql`

Creates:
- `acceptance_contract_templates`
- `acceptance_contract_template_versions`
- `task_acceptance_contracts`

Core constraints:
- unique `(template_id, version)` for immutable versioning
- `contract_source` check (`inline`, `template_ref`, `template_ref_with_overrides`)
- `validation_tier` check (`fast`, `standard`, `high_assurance`)
- foreign key from task snapshots to `tasks`

### 006 `validation_pipeline_core`

Files:
- `migrations/006_validation_pipeline_core.up.sql`
- `migrations/006_validation_pipeline_core.down.sql`

Creates:
- `critic_registry_entries`
- `validation_runs`
- `validation_stage_results`
- `validation_escalations`

Core constraints:
- `stage_name` check (`plan`, `execution`, `artifact`, `output`, `policy`, `security`, `benchmark`, `reduction`)
- `decision` check (`accepted`, `rejected`, `needs_review`)
- `status` checks for run lifecycle
- uniqueness for `(validation_run_id, stage_order, critic_id, retry_count)`

### 007 `validation_indexes_and_read_models`

Files:
- `migrations/007_validation_indexes_and_read_models.up.sql`
- `migrations/007_validation_indexes_and_read_models.down.sql`

Creates read models:
- `rm_validation_run_summary`
- `rm_validation_escalation_queue`
- `rm_critic_effectiveness`

Adds:
- lookup indexes on `task_id`, `result_id`, `tenant_id`, `validation_tier`, `status`, `created_at`
- stage-result indexes by `validation_run_id`, `stage_name`, `decision`, `critic_id`
- critic registry indexes by `critic_class`, `task_family`, `enabled`

### 008 `simulation_schema_core`

Files:
- `migrations/008_simulation_schema_core.up.sql`
- `migrations/008_simulation_schema_core.down.sql`

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

### 009 `simulation_schema_runtime`

Files:
- `migrations/009_simulation_schema_runtime.up.sql`
- `migrations/009_simulation_schema_runtime.down.sql`

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

### 010 `simulation_indexes_and_isolation`

Files:
- `migrations/010_simulation_indexes_and_isolation.up.sql`
- `migrations/010_simulation_indexes_and_isolation.down.sql`

Adds:
- lookup indexes on `scenario_id`, `scenario_version_id`, `run_id`, `tenant_id`, `status`, `created_at`
- metrics/finding indexes by `run_id`, `metric_name`, `finding_type`, `severity`
- baseline indexes by `scenario_version_id`, `created_at`
- optional partial indexes for active/running runs

Isolation checks:
- tenant-scoped access predicates supported by compound indexes
- explicit separation from production table foreign-key graph

### 011 `simulation_read_models`

Files:
- `migrations/011_simulation_read_models.up.sql`
- `migrations/011_simulation_read_models.down.sql`

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

These are data migrations and should be implemented as dedicated scripts/jobs after 005-011 are live.

## Rollback Posture

- Production posture remains forward-fix.
- `.down.sql` files exist for local/dev recovery.
- High-volume indexes should use non-blocking strategies where applicable.

## Verification Checklist

- `go test ./...` still passes after schema changes and repository wiring.
- migration apply/revert smoke in local database.
- acceptance contract and validation tables remain isolated by tenant and task ownership.
- simulation tables do not appear in production read-model materializers by default.
- simulation artifacts and events remain namespace-isolated.
