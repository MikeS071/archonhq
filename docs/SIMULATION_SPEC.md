# SIMULATION_SPEC.md

## 1. Purpose

Add a built-in synthetic proving ground to ArchonHQ so policy, scheduling, verification, reduction, pricing, and reliability behavior can be evaluated before changes are trusted in production workloads.

The simulation system exists to answer questions that task-level tests cannot answer:
- does a new verifier policy reduce false accepts or just slow the queue down?
- does a scheduler change create starvation, herding, or unhealthy market concentration?
- does a pricing change improve throughput or reward spam?
- do reliability formulas penalize bad actors without suppressing legitimate new entrants?
- do bounded self-improvement loops remain aligned under adversarial pressure?

## 2. Outcome

ArchonHQ should become:
- a control plane for agent work
- a judgment layer for verifier/reducer-backed acceptance
- an economics layer for internal settlement
- and an assurance layer with replayable simulation and stress testing

Simulation is not a side project. It is a required system capability for safe scale.

## 3. Non-goals

Do not turn simulation into:
- a social platform product
- a public sandbox for arbitrary external users
- a replacement for production acceptance tests
- a replacement for human approval on sensitive workloads
- a path for simulated results to affect production ledger, tasks, or reliability directly

## 4. Architectural Position

Simulation is a fifth platform plane: the assurance plane.

It must:
- reuse core policy, scoring, verifier, reducer, and scheduler logic where feasible
- run in isolated namespaces, storage prefixes, and event subjects
- support deterministic replay by seed and scenario version
- produce findings, reports, and baselines rather than production side effects

It must not:
- write into production task, lease, result, ledger, or reliability records
- use live tenant credentials by default
- bypass approval for replaying sensitive production traces

## 5. Primary Use Cases

### 5.1 Policy Validation
- compare current policy bundle vs candidate policy bundle
- test approval automation thresholds
- test execution policy and network restrictions
- test reliability backoff and probation rules

### 5.2 Market and Scheduler Stress Tests
- node scarcity
- queue surge
- price shocks
- redundancy overuse
- specialization starvation

### 5.3 Verifier and Reducer Quality Tests
- verifier collusion
- false consensus
- reducer instability
- merge conflict storms
- low-signal benchmark gaming
- critic monoculture
- acceptance contract drift

### 5.4 Advanced Workload Safety
- bounded autosearch/self-improve reward hacking
- rollback safety failures
- benchmark overfitting
- compute burn without quality improvement

### 5.5 Incident Replay
- replay production incidents against candidate fixes
- reproduce race conditions from historical event sequences
- validate postmortem mitigations before rollout

## 6. Service Boundary

Add `services/simulation` as an orchestration service with the following responsibilities:
- scenario registry
- run planning and admission control
- event-driven run orchestration
- synthetic actor population management
- fault injection
- metric aggregation
- findings generation
- baseline comparison
- report materialization

The simulation service depends on shared policy/scoring/verifier/reducer packages but remains isolated from production workflow writes.

## 7. Core Components

### 7.1 Scenario Registry
Stores versioned scenario definitions and metadata.

Each scenario version includes:
- `scenario_id`
- `version`
- `name`
- `goal`
- `workload_mix`
- `population_model`
- `policy_bundle_refs`
- `verifier_strategy_refs`
- `reducer_strategy_refs`
- `pricing_strategy_refs`
- `scheduler_profile`
- `seed_strategy`
- `fault_injection_profile`
- `success_criteria`
- `failure_thresholds`

### 7.2 Simulation Run Planner
Validates and starts a run from:
- scenario version
- seed
- runtime mode
- scale profile
- budget/timebox
- candidate policy overrides
- candidate formula overrides
- replay dataset refs

### 7.3 Synthetic Actor Engine
Creates run-scoped actors such as:
- workers
- verifiers
- reducers
- approvers
- malicious actors
- low-skill actors
- high-reputation incumbents
- new entrants

Actor behavior must be configurable through profiles, not hardcoded branches.

### 7.4 Runtime Bridge
Supports three execution modes:
- `deterministic_stub`
- `sampled_synthetic`
- `runtime_backed`

`runtime_backed` may call real worker adapters in isolated environments for high-fidelity tests, but is not required for every run.

### 7.5 Fault Injection Engine
Injects:
- delayed heartbeats
- dropped events
- verifier bias
- reducer non-determinism
- queue spikes
- network failures
- artifact corruption
- adversarial spam bursts

### 7.6 Findings Engine
Produces structured findings such as:
- `collusion_risk`
- `queue_instability`
- `pricing_regression`
- `reliability_bias`
- `benchmark_gaming`
- `approval_escape`
- `market_concentration`

## 8. Isolation Model

Simulation must be isolated along all of these boundaries:

### 8.1 Data Isolation
- dedicated simulation tables
- dedicated simulation read models
- dedicated object store prefixes
- dedicated event families and consumers

### 8.2 Credential Isolation
- no production provider credentials by default
- no tenant secrets by default
- synthetic or explicitly approved replay datasets only

### 8.3 Network Isolation
- default deny egress
- allowlisted endpoints only for explicitly approved runtime-backed runs

### 8.4 Operator Isolation
- simulation roles and approvals distinct from production operations
- explicit confirmation for expensive or high-fidelity runs

## 9. Scenario Types Required For v1 of Simulation

### 9.1 `scheduler_starvation_v1`
Purpose:
- detect unfair dispatch
- detect capability deadlocks
- detect incumbent lock-in

### 9.2 `verifier_collusion_v1`
Purpose:
- detect verifier clusters that mutually reinforce poor results
- measure disagreement and false-pass penetration

### 9.3 `reducer_instability_v1`
Purpose:
- detect non-deterministic or order-sensitive reductions
- measure acceptance stability across reruns

### 9.4 `market_spam_attack_v1`
Purpose:
- test pricing, reserve, and reliability defenses against spam floods

### 9.5 `approval_backlog_v1`
Purpose:
- test queue growth, SLA impact, and automation thresholds

### 9.6 `research_false_consensus_v1`
Purpose:
- test extraction quorum behavior under shared bad evidence

### 9.7 `code_patch_merge_storm_v1`
Purpose:
- test merge/reduction under conflicting patch bundles and flaky tests

### 9.8 `autosearch_reward_hacking_v1`
Purpose:
- test bounded self-improvement under benchmark gaming pressure

### 9.9 `incident_replay_v1`
Purpose:
- replay archived production incident traces with candidate fixes

### 9.10 `critic_monoculture_v1`
Purpose:
- detect approval inflation when critic bundles share providers or failure-mode classes
- measure disagreement quality versus superficial agreement

### 9.11 `acceptance_contract_drift_v1`
Purpose:
- detect false accepts and false rejects caused by stale, underspecified, or overridden acceptance contracts
- validate contract snapshotting and override guardrails

## 10. Run Modes

### 10.1 Deterministic Stub
Use lightweight synthetic actors and deterministic transitions.

Use for:
- CI
- formula regression
- policy gating
- unit and contract style simulation tests

### 10.2 Sampled Synthetic
Use stochastic actor behaviors and seeded randomness.

Use for:
- Monte Carlo policy comparison
- pricing and market stress tests
- queue dynamics

### 10.3 Runtime Backed
Use isolated real adapters and benchmark harnesses.

Use for:
- high-fidelity advanced workload evaluation
- candidate runtime or tool policy validation
- enterprise assurance packs

## 11. Data Model

Add the following tables:
- `simulation_scenarios`
- `simulation_scenario_versions`
- `simulation_runs`
- `simulation_run_steps`
- `simulation_run_actors`
- `simulation_run_artifacts`
- `simulation_run_metrics`
- `simulation_findings`
- `simulation_baselines`
- `simulation_policy_snapshots`

### 11.1 `simulation_scenarios`
Stores stable scenario identity and lifecycle.

Key fields:
- `scenario_id`
- `tenant_id nullable`
- `scope` (`platform`, `tenant`)
- `name`
- `status`
- `owner_id`
- `created_at`

### 11.2 `simulation_scenario_versions`
Stores immutable scenario definitions.

Key fields:
- `scenario_version_id`
- `scenario_id`
- `version`
- `definition_json`
- `seed_strategy`
- `runtime_mode_allowlist`
- `success_criteria_json`
- `failure_thresholds_json`
- `published_at`

### 11.3 `simulation_runs`
Stores each run and its execution state.

Key fields:
- `run_id`
- `tenant_id nullable`
- `scenario_version_id`
- `status`
- `run_mode`
- `seed`
- `started_by`
- `budget_limit_jw`
- `budget_limit_usd nullable`
- `time_limit_sec`
- `candidate_policy_json`
- `candidate_formula_json`
- `baseline_ref nullable`
- `started_at`
- `finished_at nullable`

### 11.4 `simulation_run_steps`
Stores coarse-grained step/state transitions for replay.

Key fields:
- `step_id`
- `run_id`
- `step_index`
- `event_type`
- `payload_json`
- `occurred_at`

### 11.5 `simulation_run_actors`
Stores actor population and profile assignments.

Key fields:
- `run_actor_id`
- `run_id`
- `actor_type`
- `profile_id`
- `capability_json`
- `initial_reputation`
- `malicious_probability`

### 11.6 `simulation_run_metrics`
Stores metric points and rollups.

Key fields:
- `metric_id`
- `run_id`
- `metric_name`
- `dimension_json`
- `metric_value`
- `window_name`
- `captured_at`

### 11.7 `simulation_findings`
Stores generated findings and severity.

Key fields:
- `finding_id`
- `run_id`
- `finding_type`
- `severity`
- `summary`
- `evidence_json`
- `recommended_action`

### 11.8 `simulation_baselines`
Stores promoted baseline results for regression comparison.

Key fields:
- `baseline_id`
- `scenario_version_id`
- `baseline_name`
- `metric_expectations_json`
- `promoted_from_run_id`
- `created_at`

## 12. Read Models

Add:
- `rm_simulation_run_summary`
- `rm_simulation_findings`
- `rm_simulation_baseline_diffs`
- `rm_simulation_risk_heatmap`

## 13. API Surface

Add the following endpoints:

### 13.1 Scenario Registry
- `POST /v1/simulation/scenarios`
- `GET /v1/simulation/scenarios`
- `GET /v1/simulation/scenarios/{scenario_id}`
- `POST /v1/simulation/scenarios/{scenario_id}/versions`
- `POST /v1/simulation/scenarios/{scenario_id}/publish`

### 13.2 Run Management
- `POST /v1/simulation/runs`
- `GET /v1/simulation/runs`
- `GET /v1/simulation/runs/{run_id}`
- `POST /v1/simulation/runs/{run_id}/cancel`
- `GET /v1/simulation/runs/{run_id}/events`
- `GET /v1/simulation/runs/{run_id}/metrics`
- `GET /v1/simulation/runs/{run_id}/findings`
- `GET /v1/simulation/runs/{run_id}/artifacts`

### 13.3 Baselines and Comparisons
- `POST /v1/simulation/runs/{run_id}/promote-baseline`
- `GET /v1/simulation/baselines`
- `GET /v1/simulation/baselines/{baseline_id}`
- `POST /v1/simulation/compare`

### 13.4 Replay
- `POST /v1/simulation/replays`
- `GET /v1/simulation/replays/{replay_id}`

## 14. API Rules

- all simulation mutations require idempotency keys
- large scenario definitions remain versioned and immutable once published
- expensive run creation may return `202 Accepted` and run asynchronously
- replay endpoints require elevated approval when source traces originate from sensitive production tenants
- comparison endpoints must return machine-readable metric deltas and threshold verdicts

## 15. Event Model

Add event family `simulation.*`.

Required event types:
- `simulation.scenario_created`
- `simulation.scenario_version_created`
- `simulation.scenario_published`
- `simulation.run_requested`
- `simulation.run_started`
- `simulation.step_recorded`
- `simulation.metric_recorded`
- `simulation.finding_created`
- `simulation.run_completed`
- `simulation.run_failed`
- `simulation.run_cancelled`
- `simulation.baseline_promoted`
- `simulation.replay_requested`
- `simulation.replay_completed`

## 16. NATS Subject Rules

Use:
- `jw.{tenant_id}.simulation.run_started`
- `jw.{tenant_id}.simulation.metric_recorded`
- `jw.{tenant_id}.simulation.finding_created`
- `jw.{tenant_id}.simulation.run_completed`

Consumers:
- `simulation`
- `simulation-projections`
- `web-realtime`

Simulation consumers must never update production read models.

## 17. Metrics Required

### 17.1 Core Health
- run duration
- step throughput
- completion rate
- deterministic replay match rate

### 17.2 Emergent System Metrics
- verifier disagreement rate
- false accept penetration
- false reject rate
- reducer stability score
- queue amplification factor
- scheduler starvation index
- market concentration index
- approval escape rate
- spam penetration rate
- benchmark gaming score
- reward hacking score
- compute efficiency delta

### 17.3 Economics and Trust Metrics
- credited JW drift vs baseline
- reserve exhaustion rate
- reliability bias by cohort
- new-entrant suppression rate
- cost per accepted result

## 18. Findings and Gating

Every run produces:
- summary verdict
- key metric deltas
- threshold breaches
- findings list
- recommended rollout action

Rollout actions:
- `promote`
- `promote_with_guardrails`
- `hold_for_review`
- `reject`

Policy changes in the following categories must be simulation-gated before production rollout:
- verifier strategies
- reducer strategies
- reliability formula changes
- pricing formula changes
- scheduler fairness logic
- autosearch/self-improve policy changes
- approval automation threshold changes

## 19. Security Requirements

### 19.1 Input Safety
- only synthetic, anonymized, or explicitly approved replay datasets
- strict schema validation on scenario definitions
- deny arbitrary unbounded code in deterministic and sampled modes

### 19.2 Execution Safety
- run-scoped credentials only
- no production provider secrets by default
- isolated artifact prefixes
- egress deny by default
- execution budgets and timeouts mandatory

### 19.3 Data Governance
- replay of sensitive traces requires approval and redaction policy
- findings and artifacts inherit tenant visibility rules
- production incident replay metadata must be audit logged

## 20. Operational Model

### 20.1 Scheduling
- simulation runs use dedicated queues
- production work has priority over simulation work by default
- runtime-backed simulation can be preempted when budgets or capacity policies require it

### 20.2 Storage Retention
- scenario definitions retained as long-lived records
- raw step/event detail retained per policy
- high-volume metrics downsampled after retention threshold

### 20.3 Baselines
- only published scenario versions may own baselines
- baseline promotion requires explicit human approval
- baseline changes are audited and versioned

## 21. Rollout Plan

### Phase 0: Judgment Layer Completion
- implement real verifier workflows
- implement real reducer workflows
- stop relying on placeholders for `/v1/verifications` and `/v1/reductions`

### Phase 1: Simulation Foundations
- add service boundary, tables, event family, and APIs
- support deterministic stub mode
- support scenario registry and run summaries

### Phase 2: Scenario Library
- ship the required v1 scenarios
- add baseline promotion and diffing
- add simulation dashboards

### Phase 3: Policy Gates
- require simulation comparison for critical policy changes
- wire findings into release checklists and review gates

### Phase 4: Runtime-Backed Assurance
- support isolated runtime-backed runs
- support enterprise incident replay packs
- support advanced workload benchmark packs

## 22. Acceptance Criteria

Simulation is production-ready when:
- a scenario can be created, versioned, published, and replayed by seed
- a run can execute without polluting production workflow state
- findings and metric diffs can block risky policy rollouts
- baseline regression comparisons are machine-readable
- scenario artifacts, events, and metrics remain tenant-scoped and auditable
- advanced workloads can be evaluated in simulation before production enablement

## 23. Build Checklist

- [ ] Add `services/simulation` service skeleton
- [ ] Add simulation tables and migrations
- [ ] Add scenario schema definitions
- [ ] Add scenario registry APIs
- [ ] Add run orchestration APIs
- [ ] Add simulation event family and consumers
- [ ] Add deterministic stub engine
- [ ] Add baseline promotion and diffing
- [ ] Add findings generation
- [ ] Add risk metrics to observability
- [ ] Add security controls for replay isolation
- [ ] Add dashboards for simulation runs and findings
- [ ] Gate verifier/reducer/policy changes on simulation comparison

## 24. Example Scenario Definition

```json
{
  "name": "Verifier Collusion Regression",
  "goal": "Detect false accepts when verifier cohorts become correlated",
  "workload_mix": [
    {"task_family": "research.extract", "weight": 0.7},
    {"task_family": "verify.result", "weight": 0.3}
  ],
  "population_model": {
    "workers": 200,
    "verifiers": 40,
    "malicious_verifier_fraction": 0.15
  },
  "seed_strategy": {
    "mode": "fixed",
    "seed": 424242
  },
  "success_criteria": {
    "false_accept_penetration_max": 0.03,
    "verifier_disagreement_rate_min": 0.08
  },
  "failure_thresholds": {
    "queue_amplification_factor_max": 1.8
  }
}
```
