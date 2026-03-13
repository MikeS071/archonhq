# ArchonHQ Delivery Roadmap Checklist

This checklist is aligned to the delivery milestones in `CODEX_INITIAL_PROMPT.md` and `docs/CODEX_MILESTONE_PROMPTS.md`.

## Usage

- Mark items as complete by changing `[ ]` to `[x]`.
- Keep scope changes in this file so milestone drift is visible.
- Do not run build/verification steps without explicit user approval.

## M1 Foundation

- [x] Refactor `JouleWork Network` phrasing to `ArchonHQ` (while preserving `JouleWork` economics terminology where intentional)
- [x] Scaffold monorepo root structure (`apps/`, `services/`, `pkg/`, `integrations/`, `test/`, `scripts/`)
- [x] Add root workspace/config files (Go, JS workspace, linting, env templates, make targets)
- [x] Establish clean dependency boundaries from `MONOREPO_STRUCTURE.md`
- [x] Define core domain model package skeletons in `pkg/domain`
- [x] Implement runtime adapter harness boundary for future runtimes (Hermes is production runtime in v1)
- [x] Implement auth shell (Clerk integration stubs and middleware boundaries)
- [x] Wire core infra config (Postgres, NATS, Redis, S3-compatible object storage)
- [x] Structure migrations for incremental evolution beyond `000_full_schema.sql`
- [x] Add event store foundation (`event_records` write path + typed envelope)
- [x] Implement API middleware for correlation IDs and idempotency keys on mutating routes
- [x] Establish v1 guardrails in code/config/ADR (`Hermes-only` production runtime, `ledger-only` settlement, `Postgres` durable truth and `NATS` realtime fanout only)
- [x] Add baseline observability plumbing (structured logging, tracing hooks, metrics emitter, audit stream skeleton)
- [x] Produce required initial artifacts:
- [x] Repo tree
- [x] Root config files
- [x] Migration plan
- [x] Domain package
- [x] Docker compose
- [x] Package skeletons
- [x] TODO map for milestones

## M2 Core Workflows

- [x] Implement tenants and memberships flows
- [x] Implement onboarding modes (open, invite, approval, mixed)
- [x] Implement minimum RBAC roles (`platform_admin`, `tenant_admin`, `operator`, `approver`, `auditor`, `finance_viewer`, `developer`)
- [x] Implement workspaces flows and summaries
- [x] Implement nodes registration + heartbeat lifecycle
- [x] Implement signed node registration challenge and revocable node credentials flow
- [x] Implement tasks creation/feed/detail lifecycle
- [x] Support required workload families from day one (`research.extract`, `doc.section.write`, `code.patch`, `verify.result`, `reduce.merge`, `autosearch.self_improve`)
- [x] Implement approvals queue + approve/deny flows
- [x] Implement leases create/claim/release/extend flows
- [x] Persist workflow events to `event_records` for each transition
- [x] Implement API endpoint surface for core workflows under `/v1` with Clerk/node auth boundaries
- [x] Add projection/materializer read models (`rm_active_tasks`, `rm_approval_queue`, `rm_fleet_overview`, `rm_node_heartbeat`, `rm_task_trace`, `rm_ledger_balances`, `rm_reliability_summary`, `rm_recent_settlements`)
- [x] Enforce tenant isolation across all workflow paths
- [x] Use the REVIEW_GATE_PROMPT.md to review specs and code delivered and ensure there aren't any gaps in this phase

## M3 Worker Runtime

- [x] Scaffold `apps/worker-node`
- [x] Implement Hermes adapter interface in production path
- [x] Implement backend policy mapping for Docker/SSH/Modal execution
- [x] Enforce isolated ephemeral task workspaces by default
- [x] Enforce no write-back to long-lived Hermes personal memory by default
- [x] Enforce per-lease network policies and tool grants
- [x] Implement artifact upload/register/download flow
- [x] Enforce artifact persistence by object storage reference only (no large artifact bytes in Postgres)
- [x] Implement signed result submission and verification hooks
- [x] Capture run telemetry (logs, tool calls, metrics) and persist references
- [x] Enforce BYOK inference-only runtime behavior for v1
- [x] Use the REVIEW_GATE_PROMPT.md to review specs and code delivered and ensure there aren't any gaps in this phase

## M4 Economics

- [x] Implement raw JouleWork computation
- [x] Implement quality scoring pipeline
- [x] Implement reliability snapshot model and update flow
- [x] Implement pricing quote and rate resolution strategies
- [x] Implement settlement engine and ledger posting
- [x] Implement reserve hold creation/release lifecycle
- [x] Expose operator earnings and reserve summaries
- [x] Use the REVIEW_GATE_PROMPT.md to review specs and code delivered and ensure there aren't any gaps in this phase

## M5 UI (Svelte/SvelteKit + shadcn-svelte)

- [x] For frontend UI use the shadcn-svelte kit (https://www.shadcn-svelte.com/llms.txt)
- [x] Scaffold Svelte + shadcn-svelte app routes from `frontend/FRONTEND_ROUTE_COMPONENT_MAP.md`
- [x] Implement dashboard with key metric cards
- [x] Implement tasks list and task detail tabs
- [x] Implement approvals queue UI
- [x] Implement fleet and node detail UI
- [x] Implement ledger and reliability pages
- [x] Implement pricing and provider settings pages
- [x] Implement admin surfaces and role-aware guards
- [x] Use the REVIEW_GATE_PROMPT.md to review specs and code delivered and ensure there aren't any gaps in this phase
  - Current review report: `docs/reviews/M5_REVIEW_GATE.md` (M5 gaps closed; gate complete)

## M6 Integrations

- [ ] Implement Paperclip connector service boundary
- [ ] Sync workspace summary projections to Paperclip surfaces
- [ ] Sync approval queue and ticket/task projection state
- [ ] Sync fleet heartbeat summaries
- [ ] Sync settlement/reliability projection metrics
- [ ] Ensure Paperclip is never used as durable source of truth
- [ ] Use the REVIEW_GATE_PROMPT.md to review specs and code delivered and ensure there aren't any gaps in this phase

## M7 Advanced Workloads

- [ ] Implement code patch merge flow strategies
- [ ] Implement bounded autoresearch/self-improve workflow loop
- [ ] Add guardrails for iteration limits, budget, and approval gates
- [ ] Add evaluator/verifier hooks for iterative workloads
- [ ] Add auditable experiment/result lineage views
- [ ] Add simulation entry points for advanced workload policy and benchmark testing
- [ ] Use the REVIEW_GATE_PROMPT.md to review specs and code delivered and ensure there aren't any gaps in this phase

## M8 Simulation and Assurance

- [x] Complete M8 prep artifacts (`docs/SIMULATION_SPEC.md`, `docs/openapi/openapi.yaml`, `docs/M8_MIGRATION_SPEC.md`, `services/simulation/README.md`)
- [ ] Add `services/simulation` service boundary and scenario registry
- [ ] Implement dedicated simulation tables, read models, and event family
- [ ] Implement replayable event-driven runs with fixed-seed support
- [ ] Implement deterministic stub mode for CI and policy regression
- [ ] Implement sampled synthetic mode for market and queue stress tests
- [ ] Add required v1 scenarios (`scheduler_starvation_v1`, `verifier_collusion_v1`, `reducer_instability_v1`, `market_spam_attack_v1`, `approval_backlog_v1`, `research_false_consensus_v1`, `code_patch_merge_storm_v1`, `autosearch_reward_hacking_v1`, `incident_replay_v1`)
- [ ] Implement baseline promotion and metric diffing
- [ ] Implement findings generation and risk heatmaps
- [ ] Add simulation dashboards and operator views
- [ ] Gate verifier/reducer/scheduler/pricing/reliability policy changes on simulation comparison
- [ ] Use the REVIEW_GATE_PROMPT.md to review specs and code delivered and ensure there aren't any gaps in this phase

## Cross-Cutting Quality Gates

- [ ] API contracts aligned to `docs/API_CONTRACTS.md` and `docs/openapi/openapi.yaml`
- [x] Mutating API routes enforce idempotency-key semantics and emit correlation IDs
- [x] Error envelope and codes aligned to `docs/ERROR_MODEL.md`
- [ ] Security controls aligned to `docs/SECURITY_MODEL.md`
- [ ] Secrets encrypted, tenant-scoped credentials enforced, and object storage namespace isolation validated
- [ ] Observability coverage aligned to `docs/OBSERVABILITY_SPEC.md`
- [ ] Key latency metrics implemented (`approval`, `lease`, `result submission`, `verification`, `reduction`, `settlement`)
- [ ] Emergent-risk metrics implemented (`verifier disagreement`, `false accept penetration`, `reducer stability`, `queue amplification`, `scheduler starvation`, `market concentration`, `approval escape`)
- [ ] Policy model aligned to `docs/POLICY_SCHEMA.md`
- [ ] NATS subjects and consumer groups aligned to `docs/NATS_SUBJECT_MAP.md`
- [ ] Sequence flow fidelity aligned to `docs/SEQUENCE_DIAGRAMS.md`
- [x] Test coverage progression aligned to `docs/TEST_PLAN.md`
- [ ] Contract tests present for Hermes adapter and Paperclip connector
- [ ] Contract tests present for simulation registry and run APIs
- [x] Security tests present for tenant isolation, forbidden access checks, and invalid signature rejection
- [ ] Replay approval enforcement and simulation namespace isolation validated
- [ ] Use the REVIEW_GATE_PROMPT.md to review specs and code delivered and ensure there aren't any gaps in this phase

## Milestone Exit Criteria

- [x] M1 exit: foundation scaffolding complete and runnable core wiring in place
- [x] M2 exit: task lifecycle from tenant/workspace through lease + events/projections working
- [x] M3 exit: worker runtime can execute, upload artifacts, and submit signed results
- [x] M4 exit: scoring, pricing, ledger, and reserve flows operational
- [x] M5 exit: Svelte operator workflows available for core operations
- [ ] M6 exit: Paperclip projections syncing from internal source-of-truth state
- [ ] M7 exit: advanced merge and bounded self-improvement flows operational
- [ ] M8 exit: simulation scenarios, replayable runs, baseline comparisons, and policy gates operational
- [ ] Use the REVIEW_GATE_PROMPT.md to review specs and code delivered and ensure there aren't any gaps in this phase
