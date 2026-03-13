# Milestone TODO Map

## M1 Foundation

- Monorepo scaffolding and root config files
- Core package boundaries and domain models
- Auth shell and middleware boundaries
- Infra wiring (Postgres, NATS, Redis, object storage)
- Event envelope/store foundation

## M2 Core Workflows

- Tenants/workspaces/nodes/tasks/approvals/leases APIs
- Event writes for workflow transitions
- Materialized read models and tenant isolation

## M3 Worker Runtime

- Worker-node process
- Hermes adapter implementation
- Docker/SSH/Modal execution policy mapping
- Artifact and signed result flow

## M4 Economics

- Raw JouleWork and quality/reliability scoring
- Pricing quote/rate resolution
- Settlement posting and reserve lifecycle

## M5 UI

- Svelte + shadcn-svelte routes and dashboard pages
- Task trace, approvals, fleet, ledger, reliability

## M6 Integrations

- Paperclip connector for projected workflow state

## M7 Advanced Workloads

- Code patch merge strategies
- Bounded autosearch self-improve loops
- Evaluator/verifier hooks for iterative workloads
- Auditable experiment and result lineage

## M8 Simulation and Assurance

- Prep artifacts complete (`docs/M8_PREP_WORK.md`, `docs/M8_MIGRATION_SPEC.md`, `docs/openapi/openapi.yaml`)
- Acceptance contract templates and task snapshot model
- Validation tier routing and critic diversity policy
- Critic registry and stage-gated validation orchestration
- Simulation service and scenario registry
- Replayable event-driven runs
- Baseline promotion and diffing
- Emergent-risk metrics and findings
- Policy rollout gates for critical changes, including validation-policy widening

## M9 Open Market Network

- Market profiles for requesters and executors
- Work classes and publication controls
- Listing publication, bids, claims, and shard market flows
- Funded reserve checks and task escrow lifecycle
- Payout accounts and payout request flow
- Dispute and arbitration service boundaries
- Market reputation and anti-abuse controls
- Market-mode simulation scenarios and rollout gates
