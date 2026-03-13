# Reliability and Assurance Build Spec

## 1. Purpose

This document turns the current ArchonHQ direction into an implementation-ready build spec for the next platform hardening phase.

The directional adjustment is:

- ArchonHQ remains a broad agent-work platform.
- The platform should more explicitly optimize for trust, judgment, and assurance rather than generic orchestration alone.
- Reliability should come from pre-declared acceptance contracts, stage-gated critics with veto authority, controlled execution boundaries, and simulation-gated rollout of automation.

This spec is intended to drive M8 implementation and adjacent cross-cutting work.

Open-market mode is specified separately in `docs/OPEN_MARKET_NETWORK_BUILD_SPEC.md` and depends on the M8 judgment-layer work completing first.

## 2. Outcome

ArchonHQ should behave like a production judgment layer for distributed agent work:

- tasks declare what success means before execution starts
- work does not advance because a producer says it is done
- independent critics evaluate stage outputs against explicit gates
- validation depth is selected by policy and risk tier
- critical policy changes are simulation-gated before rollout

This is a strengthening of current platform intent, not a product pivot.

## 3. Non-goals

- Do not narrow the platform to a single vertical workflow or domain.
- Do not require every task to use the heaviest validation path.
- Do not treat critic count as a goal by itself.
- Do not replace typed reduction with majority voting.
- Do not claim near-perfect automation where the residual error floor is still material.

## 4. Product Adjustment

ArchonHQ should keep its existing major operational planes while treating judgment as a first-class cross-cutting layer:

1. control plane for tasking and policy
2. execution plane for isolated worker runtime behavior
3. operator plane for approvals and visibility
4. storage plane for durable truth, artifacts, and projections
5. assurance plane for simulation, replay, baselines, and rollout gates

Cross-cutting capability layers:

- judgment layer for acceptance contracts, critics, verification, and reduction
- economics layer for pricing, reliability, and settlement

The main adjustment is that the judgment layer becomes an explicit implementation target rather than an emergent property of existing verification and reduction services.

## 5. Capability Additions

### 5.1 Acceptance Contracts

Every production task must carry an acceptance contract either inline or via template reference.

The acceptance contract defines:

- `contract_id`
- `contract_version`
- `name`
- `task_family`
- `validation_tier`
- `objectives`
- `required_evidence`
- `required_checks`
- `scoring_rules`
- `pass_thresholds`
- `needs_review_thresholds`
- `max_stage_retries`
- `escalation_rules`
- `required_critic_classes`
- `allow_self_certification` and default `false`

The contract is immutable once attached to an approved task.

The contract must be snapshot-copied onto the task so later template changes do not alter in-flight or completed audit history.

Acceptance contracts should support the following modes:

- `template_ref`
- `inline`
- `template_ref_with_overrides`

Overrides must be explicit, audit logged, and policy checked.

### 5.2 Critic Registry

Add a registry of critic definitions that can be selected by family, stage, and validation tier.

Each critic definition includes:

- `critic_id`
- `version`
- `stage`
- `critic_class`
- `task_families`
- `validation_tiers`
- `failure_mode_class`
- `provider_requirements`
- `input_contract`
- `decision_contract`
- `default_thresholds`
- `max_retries_before_escalation`
- `enabled`

Required stage values:

- `plan`
- `execution`
- `artifact`
- `output`
- `policy`
- `security`
- `benchmark`

Required critic classes for initial rollout:

- `plan_soundness`
- `evidence_completeness`
- `artifact_integrity`
- `output_correctness`
- `policy_compliance`
- `security_compliance`
- `benchmark_regression`

The registry is configuration and policy driven. It must not require code edits for routine activation or deactivation of critic bundles.

### 5.3 Stage-Gated Validation Pipeline

Validation should be modeled as an orchestration pipeline instead of a single generic verification threshold.

The minimum production pipeline is:

1. plan gate
2. execution gate
3. artifact or evidence gate
4. output gate
5. policy or security gate when required by family or sensitivity
6. reduction gate when multiple results compete or merge

Each stage produces:

- `decision`: `accepted`, `rejected`, `needs_review`
- `score`
- `critic_outputs`
- `evidence_refs`
- `retry_recommended`
- `escalation_required`

Rejection returns authority to the producing stage or team rather than auto-advancing to later stages.

Producers cannot approve their own work for stage completion.

### 5.4 Validation Tiers

Validation depth must be selectable by policy.

Define three tiers:

- `fast`
- `standard`
- `high_assurance`

Tier expectations:

`fast`
- low-risk or exploratory tasks
- minimal critic set
- lower retry budget
- may skip certain holistic critics

`standard`
- default production mode
- required for most tenant workloads
- includes plan, output, and policy-aware checks where applicable

`high_assurance`
- sensitive or high-consequence tasks
- full critic bundle with stronger evidence completeness requirements
- stricter escalation and human-review triggers
- simulation-gated policy changes before widening automation

Policy decides the maximum allowed tier downgrade for a given family, sensitivity tier, tenant, or workspace.

### 5.5 Critic Diversity Rules

Critic diversity matters more than critic count.

Add policy-enforceable diversity constraints:

- prohibit all required critics for a stage bundle from sharing the same `failure_mode_class`
- optionally require cross-provider diversity for selected families or tiers
- optionally require critic and producer separation by provider or model family

These constraints should be soft in `standard` mode and hard in `high_assurance` mode.

### 5.6 Tightened Brains vs Hands Separation

Strengthen the runtime contract between reasoning and execution:

- agents should consume summaries, schemas, sample rows, typed extracts, and evidence refs by default
- raw tool output should remain in execution context unless explicitly allowlisted
- critics should evaluate artifacts, evidence, and structured outputs rather than unbounded transcripts where possible
- sensitive datasets and large tool responses should remain outside model context unless policy explicitly permits access

This expands existing workspace isolation into a clearer reasoning/execution boundary.

## 6. Data Model

### 6.1 New Core Records

Add the following logical records:

- `acceptance_contract_templates`
- `acceptance_contract_template_versions`
- `task_acceptance_contracts`
- `critic_registry_entries`
- `validation_runs`
- `validation_stage_results`
- `validation_escalations`

### 6.2 `task_acceptance_contracts`

Required fields:

- `task_id`
- `tenant_id`
- `contract_source`
- `template_id nullable`
- `template_version nullable`
- `contract_json`
- `validation_tier`
- `attached_at`
- `attached_by`

### 6.3 `validation_runs`

Required fields:

- `validation_run_id`
- `tenant_id`
- `workspace_id nullable`
- `task_id`
- `result_id nullable`
- `reduction_id nullable`
- `status`
- `validation_tier`
- `acceptance_contract_snapshot`
- `started_at`
- `completed_at nullable`
- `stop_reason nullable`

### 6.4 `validation_stage_results`

Required fields:

- `validation_run_id`
- `stage_name`
- `stage_order`
- `critic_id`
- `critic_version`
- `decision`
- `score`
- `failure_mode_class`
- `provider_id nullable`
- `evidence_refs_json`
- `details_json`
- `retry_count`
- `created_at`

## 7. API Build Spec

### 7.1 Task Create and Read

Extend `POST /v1/tasks` and task detail responses with:

- `acceptance_contract`
- `acceptance_contract_template_id nullable`
- `validation_tier`

Create requests must reject tasks that:

- omit both inline contract and template reference when the family requires one
- request a tier forbidden by policy
- attempt to enable self-certification on a disallowed family

### 7.2 Acceptance Contract Templates

Add endpoint group:

- `POST /v1/acceptance-contract-templates`
- `GET /v1/acceptance-contract-templates`
- `GET /v1/acceptance-contract-templates/{template_id}`
- `POST /v1/acceptance-contract-templates/{template_id}/versions`
- `POST /v1/acceptance-contract-templates/{template_id}/publish`

Rules:

- versions are immutable after publish
- publish requires tenant-admin or higher privileges
- template use on high-assurance families requires policy compatibility checks

### 7.3 Critic Registry

Add endpoint group:

- `GET /v1/critics`
- `GET /v1/critics/{critic_id}`
- `POST /v1/critics`
- `POST /v1/critics/{critic_id}/versions`
- `POST /v1/critics/{critic_id}/publish`

In production, write access should be tenant- or platform-admin restricted.

### 7.4 Validation Runs

Add endpoint group:

- `POST /v1/tasks/{task_id}/validation-runs`
- `GET /v1/tasks/{task_id}/validation-runs`
- `GET /v1/validation-runs/{validation_run_id}`
- `GET /v1/validation-runs/{validation_run_id}/stages`
- `POST /v1/validation-runs/{validation_run_id}/escalate`

Rules:

- create may return `202 Accepted`
- each run snapshots the acceptance contract and critic bundle selection
- stage results are append-only

### 7.5 Validation Bundles in Existing Endpoints

Existing verification and reduction endpoints must return:

- `validation_run_id`
- `acceptance_contract_ref`
- `validation_tier`
- `stage_summary`

### 7.6 Response Shape

Task detail should include:

```json
{
  "task_id": "task_01",
  "validation_tier": "standard",
  "acceptance_contract": {
    "contract_id": "ac_01",
    "contract_version": 2,
    "required_critic_classes": [
      "plan_soundness",
      "evidence_completeness",
      "output_correctness"
    ]
  },
  "latest_validation_run": {
    "validation_run_id": "vr_01",
    "status": "needs_review",
    "failed_stage": "output"
  }
}
```

## 8. Event Model

Add event families:

- `acceptance_contract.*`
- `validation.*`

Required events:

- `acceptance_contract.template_created`
- `acceptance_contract.version_published`
- `task.acceptance_contract_attached`
- `validation.run_started`
- `validation.stage_passed`
- `validation.stage_rejected`
- `validation.stage_needs_review`
- `validation.escalated`
- `validation.run_completed`

Validation events must include the selected critic identity, failure-mode class, and evidence refs used for the decision.

## 9. Policy Model

Add a `validation` policy section with:

- `default_tier`
- `family_tier_overrides`
- `require_acceptance_contract`
- `allow_inline_contracts`
- `allow_contract_overrides`
- `critic_diversity_mode`
- `require_cross_provider_critics`
- `max_stage_retries`
- `human_review_rules`

The `security` policy section should also support:

- `allow_raw_tool_output_to_models`
- `restricted_evidence_types`
- `require_evidence_redaction_for_validation`

## 10. Observability Build Spec

Required new metrics:

- acceptance contract coverage by family
- validation runs by tier
- stage retry depth
- critic catch contribution by class
- evidence missing rate
- escalation residual rate
- critic monoculture ratio
- contract override frequency
- validation latency by stage
- validation cost by tier

Required dashboards:

- validation pipeline health
- critic effectiveness and drift
- tier routing distribution
- human escalation backlog
- simulation gate results for candidate policy changes

## 11. Security Build Spec

Required controls:

- acceptance contract snapshots are immutable after task approval
- critic registry writes are audit logged
- validation runs must obey tenant isolation
- critic evidence access must be least privilege
- raw tool output exposure to models must be policy gated and logged
- cross-tenant critic reuse must not imply cross-tenant data visibility
- high-assurance validation overrides require explicit approval

## 12. Simulation Build Spec

M8 should validate not only scheduler or reducer policy changes, but also judgment-layer changes.

Add required simulation scenarios:

- `critic_monoculture_v1`
- `acceptance_contract_drift_v1`

Intent:

`critic_monoculture_v1`
- measure degradation when all critics share the same provider or failure-mode class
- surface agreement without real coverage

`acceptance_contract_drift_v1`
- measure false accepts and false rejects when acceptance rules are underspecified, stale, or partially overridden

These scenarios complement the existing list and are required before broad automation widening for validation-heavy workloads.

## 13. Implementation Order

### Phase A

- acceptance contract templates and snapshots
- validation tier policy
- task API changes
- event and audit changes

### Phase B

- critic registry
- stage-gated validation run orchestration
- updated verification and reduction output contracts
- critic effectiveness telemetry

### Phase C

- simulation scenarios for judgment-layer regressions
- rollout gates tied to simulation compare results
- operator dashboards for contract, critic, and escalation visibility

## 14. Exit Criteria

The reliability hardening work is complete when:

- every production family has an acceptance contract path
- validation tier routing is policy enforced
- stage-gated validation runs are operational for task families that require verification or reduction
- critic diversity controls are enforced for high-assurance workflows
- simulation covers critic monoculture and acceptance contract drift
- policy changes affecting validation cannot widen automation without passing simulation comparison
