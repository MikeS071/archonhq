# SPEC.md — ArchonHQ Complete Product and Implementation Specification

## 1. Mission

Build **ArchonHQ**, a multi-tenant platform for coordinating, approving, executing, verifying, reducing, scoring, and internally settling work performed by agent-operated client nodes.

The system must support:
- broad workload classes
- sensitive/private data handling
- human approval by default
- Hermes-powered worker nodes
- Paperclip-dependent operator workflow
- a built-in synthetic proving ground for pre-production assurance
- ledger-only settlement in v1
- clear operator UX for tasks, reliability, and earnings

## 2. Product overview

### 2.1 Primary actor types
- Platform admin
- Tenant admin
- Operator
- Approver
- Auditor
- Node
- Agent
- Verifier
- Reducer

### 2.2 Major planes
- Control plane
- Execution plane
- Operator plane
- Storage plane
- Assurance plane

### 2.3 Major technical choices
- Backend: Go
- Frontend: Svelte / SvelteKit / shadcn-svelte
- Human auth: Clerk
- Durable store: Postgres
- Event fanout/realtime: NATS
- Cache: Redis
- Artifacts: S3-compatible object storage
- Worker runtime: Hermes Agent
- Operator workflow dependency: Paperclip

## 3. Fixed decisions from the conversation

1. Onboarding supports both open signup and permissioned/approval modes.
2. v1 is internal accounting only; no external payouts.
3. Compliance posture is global best-effort.
4. Inference is BYOK only in v1.
5. Broad workload support is required from day one.
6. Required workload families:
   - research.extract
   - doc.section.write
   - code.patch
   - verify.result
   - reduce.merge
   - autosearch.self_improve
7. Multi-tenant from day one.
8. Paperclip is a required dependency in v1.
9. Hermes is the only production runtime in v1.
10. Adapter harness must exist for future runtimes.
11. Docker, SSH, Modal are required backends.
12. Frontend uses Svelte + shadcn-svelte.
13. Durable truth uses Postgres; NATS is fanout/realtime only.
14. Human auth uses Clerk.
15. Sensitive/private data is supported by default.
16. Marketplace workspace memory is isolated and ephemeral by default.
17. Human approval is the default execution mode.
18. Pricing supports both fixed rate cards and dynamic bidding.
19. Optimize for clean long-term architecture and good UX.
20. Built-in simulation/proving-ground support is required before critical policy automation is widened.
21. Simulation state must remain isolated from production workflow truth.

## 4. Problem statement

This system exists because distributed agent work does not fit cleanly into Git/GitHub-style collaboration alone:
- DHT sharding solves distribution, not merge semantics
- line diffs are not enough for structured extractions and reductions
- agent work needs approvals, verification, reducer logic, trust scoring, and accounting
- operators need direct visibility into what their agents are doing and what value they have produced
- system-level behavior under scale and adversarial pressure must be measured before changes are trusted

## 5. External reference projects and how they influence this design

### 5.1 Paperclip
Use as:
- operator workflow dependency
- governance/approval UX inspiration and integration target
- ticket-style work projection surface
- budget and heartbeat projection surface

Do not use as:
- source of truth for tasks
- source of truth for leases
- source of truth for ledger/reliability/events

### 5.2 Hermes Agent
Use as:
- v1 worker runtime
- execution adapter target
- tool and backend execution layer
- MCP-aware extension boundary

### 5.3 SporeMesh
Use as conceptual inspiration for:
- simple node joining
- immutable signed work records
- low-friction participation

### 5.4 Karpathy-related autoresearch context
Use as inspiration for:
- bounded experiment search loops
- reproducible benchmark-driven self-improvement workloads
- auditable reducer/verifier-friendly iteration

## 6. Architectural principles

1. Immutable artifacts
2. Durable append-only events
3. No direct worker mutation of canonical accepted state
4. Reducers/materializers create accepted state
5. Typed merge strategies
6. Typed verifier strategies
7. Approval-first by default
8. Sensitive-data-safe defaults
9. Explicit policy snapshots
10. Ledger-first economics
11. Version all schemas and strategies
12. Keep service boundaries clean
13. Reuse business logic across production and simulation where feasible
14. Never mix simulation writes with production workflow truth

## 7. Major product features

### 7.1 Tenancy
- tenants
- memberships
- roles
- workspace isolation
- tenant-level policies
- open / invite / approval / mixed onboarding

### 7.2 Marketplace
- task creation
- task discovery/feed
- approval gates
- fixed-price or bid-based execution
- redundancy/competition
- verifier and reducer tasks
- recursive decomposition

### 7.3 Worker node runtime
- node registration
- capability reporting
- lease polling/claiming
- isolated workspace creation
- Hermes execution
- artifact upload/download
- signed result submission
- heartbeats and telemetry

### 7.4 Operator UX
- dashboard
- approval queue
- fleet overview
- task trace detail
- reliability dashboard
- ledger/earnings dashboard
- provider key management
- policy management

### 7.5 Economics
- raw JouleWork
- quality scoring
- reliability scoring
- credited JouleWork
- pricing resolution
- internal ledger posting
- reserve holds

### 7.6 Assurance and simulation
- versioned scenario registry
- replayable simulation runs
- synthetic and incident-replay modes
- policy and formula comparison
- emergent-risk findings
- baseline promotion
- rollout gates for high-impact changes

## 8. Roles and access model

Minimum roles:
- platform_admin
- tenant_admin
- operator
- approver
- auditor
- finance_viewer
- developer

## 9. Identity model

### 9.1 Humans
Use Clerk for:
- authentication
- tenant membership linkage
- sessions
- JWT for API calls

### 9.2 Nodes
Each node has:
- keypair
- node_id
- signed registration challenge
- revocable credential

### 9.3 Agents
Each runtime agent has:
- agent_id
- node linkage
- lineage metadata
- task/workspace linkage

## 10. Security model summary

The default assumption is that tasks may contain private or sensitive enterprise data.

Requirements:
- TLS everywhere
- secret vaulting
- signed node auth
- signed result submission
- object storage isolation
- audit logs
- policy-controlled tool grants
- policy-controlled network access
- isolated ephemeral marketplace workspaces
- no write-back to long-lived Hermes personal memory by default

See `docs/SECURITY_MODEL.md`.

## 11. Workload family registry

### 11.1 research.extract
Input:
- URLs, documents, extraction schema

Output:
- structured records, evidence refs

Merge:
- quorum_fact_v1, append_only_v1, key_upsert_v1

Verifier:
- schema, factuality, duplication

### 11.2 doc.section.write
Input:
- outline, evidence pack, style guide

Output:
- section artifacts, section ops

Merge:
- section_patch_v1, topk_rank_v1

Verifier:
- style, coherence, evidence completeness

### 11.3 code.patch
Input:
- repo bundle, target state, tests, policy bundle

Output:
- patch bundles, AST ops, test logs

Merge:
- ast_patch_v1

Verifier:
- compile, tests, policy, security

### 11.4 verify.result
Input:
- task + candidate result

Output:
- verification report

### 11.5 reduce.merge
Input:
- candidate results

Output:
- accepted reduction or merged state

### 11.6 autosearch.self_improve
Input:
- benchmark harness
- bounded search space
- code/config basis

Output:
- candidate code/config changes
- benchmark artifacts
- reducer/verifier-ready result bundles

## 12. Domain entities

- Tenant
- Membership
- Operator
- Workspace
- Node
- Agent
- ProviderCredential
- PolicyBundle
- TaskSpec
- Lease
- ApprovalRequest
- Artifact
- ResultClaim
- VerificationReport
- Reduction
- ReliabilitySnapshot
- PriceQuote
- RateSnapshot
- LedgerAccount
- LedgerEntry
- ReserveHold
- EventRecord
- ProjectionState
- SimulationScenario
- SimulationRun
- SimulationFinding
- SimulationBaseline

## 13. Canonical formulas

### 13.1 JouleWork
raw_jw = cpu_component + gpu_component + token_component + tool_component + io_component

cpu_component = cpu_sec * 0.002
gpu_component = gpu_sec * gpu_weight(gpu_class)
token_component = (tokens_in + 2*tokens_out) * 0.000002
tool_component = external_tool_calls * 0.02
io_component = network_mb * 0.0005 + storage_mb * 0.0002
raw_jw_final = raw_jw * task_multiplier

GPU weights:
- cpu-only = 0.010
- mps = 0.018
- rtx_3060_class = 0.025
- rtx_4090_class = 0.050
- a100_h100_class = 0.090

Task multipliers:
- easy = 0.8
- standard = 1.0
- hard = 1.25
- critical = 1.6

All coefficients configurable by policy.

### 13.2 Quality
Q = 0.35*validity + 0.30*verifier_score + 0.20*acceptance_signal + 0.10*novelty + 0.05*latency_score

Autosearch quality:
Q_autosearch = 0.30*benchmark_delta_norm + 0.25*eval_reproducibility + 0.20*rollback_safety + 0.15*search_novelty + 0.10*compute_efficiency

### 13.3 Reliability
agent_rf = 0.30*validity_rate + 0.25*verification_pass_rate + 0.20*acceptance_rate + 0.15*(1 - rollback_rate) + 0.10*(1 - dispute_loss_rate)

operator_rf = 0.40*fleet_acceptance_rate + 0.25*fleet_verification_rate + 0.15*(1 - fleet_rework_rate) + 0.10*(1 - dispute_rate) + 0.10*uptime_score

effective_rf = 0.65*agent_rf + 0.35*operator_rf
rf_final = 0.50*rf_last_100 + 0.30*rf_last_30d + 0.20*rf_lifetime

### 13.4 Credited JouleWork
credited_jw = raw_jw_final * quality_factor * reward_multiplier_from_rf

RF reward multipliers:
- >=0.95 => 1.00
- 0.90–0.949 => 0.92
- 0.80–0.899 => 0.75
- 0.70–0.799 => 0.50
- 0.55–0.699 => 0.20
- <0.55 => 0.00

### 13.5 Budget split defaults
- 15% participation pool
- 65% acceptance pool
- 20% reliability dividend

## 14. Approval model

Approval modes:
- always_required
- tenant_default
- operator_optional
- fully_automated
- risk_based

Approval scopes:
- tenant
- workspace
- family
- task
- node
- agent
- provider/model
- toolset
- backend
- spend threshold
- sensitivity class

Approval-first is default.

## 15. Scheduler and leasing

Scheduler inputs:
- capability match
- reliability
- sensitivity policy
- backend availability
- approval state
- price/budget
- redundancy settings

Lease modes:
- exclusive
- redundant
- verifier
- reducer
- shard
- recursive

Reliability-aware leasing defaults:
- >= 0.95 => 64 active leases
- 0.90–0.949 => 32
- 0.80–0.899 => 12
- 0.70–0.799 => 4
- 0.55–0.699 => 1
- < 0.55 => probation-only

## 16. Artifact model

Artifacts are:
- immutable
- content-addressed
- encrypted at rest
- stored in S3-compatible storage
- described by metadata in Postgres

Types:
- input bundles
- output bundles
- traces
- logs
- screenshots
- benchmark artifacts
- verifier notes
- reduced outputs

## 17. Merge/reduction model

Required merge strategies:
- append_only_v1
- key_upsert_v1
- section_patch_v1
- ast_patch_v1
- topk_rank_v1
- quorum_fact_v1
- reduce_tree_v1

Reducers and materializers are the only path to canonical accepted state.

## 18. Verification model

Required verifier types:
- schema
- policy
- compile
- tests
- factual/quorum
- semantic coherence
- benchmark/eval
- duplication/fraud
- simulation/baseline-regression

## 19. Operator UI requirements

Routes:
- /dashboard
- /tasks
- /tasks/[task_id]
- /approvals
- /fleet
- /fleet/nodes/[node_id]
- /ledger
- /reliability
- /pricing
- /settings/providers
- /admin

Questions the UI must answer:
- what are my agents doing?
- what needs approval?
- what got accepted or rejected and why?
- how much raw JouleWork did I produce?
- how much credited JouleWork did I earn?
- what is my effective RF?
- why did my payout-like ledger differ from raw effort?

## 20. Backend service requirements

Services:
- api gateway
- scheduler
- approvals
- verification
- reduction
- reliability
- simulation
- joulework
- pricing
- ledger
- notifications
- paperclip connector

## 21. Infra requirements

- local docker-compose
- Kubernetes-friendly deployment structure
- migrations
- OpenTelemetry
- Redis for cache only
- NATS for realtime only
- Postgres for durable truth
- isolated simulation object-store prefixes and consumer groups

## 22. Testing requirements

- unit tests for formulas
- integration tests for task -> approval -> lease -> result -> verification -> reduction -> settlement
- tenancy isolation tests
- worker runtime contract tests
- Paperclip connector contract tests
- end-to-end local smoke flow
- deterministic simulation regression tests
- scenario baseline comparison tests
- incident replay safety tests

## 23. Exclusions for v1

Do not build in v1:
- real money payouts
- blockchain/token settlement
- p2p control plane
- production runtimes beyond Hermes
- unrestricted anonymous execution

## 24. Build acceptance criteria

1. tenant creation works
2. workspace creation works
3. node registration works
4. Hermes-backed Docker/SSH/Modal execution works
5. approval -> lease -> execute -> verify -> reduce -> settle works
6. artifacts land in object store
7. events persist to Postgres and fan out through NATS
8. UI shows tasks, approvals, fleet, reliability, and ledger
9. Paperclip workflow projection exists
10. isolated ephemeral task workspaces are the default
11. simulation scenarios can be created, run, replayed, and compared without affecting production workflow records
