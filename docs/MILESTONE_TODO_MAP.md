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
