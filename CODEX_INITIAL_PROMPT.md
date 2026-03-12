# CODEX_INITIAL_PROMPT.md

Build this project named **archonhq** with github repo at (https://github.com/MikeS071/archonhq.git). The local directory at ~/projects/archonhq is upto date.

Refactor any "JouleWork Network" phrasing to "ArchonHQ"

Use this kit as the source of truth:
- SPEC.md
- MONOREPO_STRUCTURE.md
- docs/API_CONTRACTS.md
- docs/DB_SCHEMA_DRAFT.md
- docs/EVENT_CATALOG.md
- docs/NATS_SUBJECT_MAP.md
- docs/POLICY_SCHEMA.md
- docs/HERMES_ADAPTER_SPEC.md
- docs/PAPERCLIP_CONNECTOR_SPEC.md
- docs/TEST_PLAN.md
- docs/ERROR_MODEL.md
- docs/OBSERVABILITY_SPEC.md
- docs/SECURITY_MODEL.md
- docs/SEQUENCE_DIAGRAMS.md
- docs/UI_WIREFRAMES.md
- docs/go_interfaces - contains go interface specs
- docs/openapi - contains api specs
- docs/schemas - contains db schema specs

Examples of key code artefacts exist in the following directories:
- **examples/** - use those examples to kickstart the build process.
- **frontend/** - front end routing components
- **migrations/** - db migration scripts and stubs

## Phased implementation approach

Follow the phased implementation approach below:

### M1 Foundation
Scaffold monorepo, root files, config, migrations, domain models, auth shell, infra wiring.

### M2 Core workflows
Implement tenants, workspaces, nodes, tasks, approvals, leases, event store, projections.

### M3 Worker runtime
Implement worker-node, Hermes adapter, Docker/SSH/Modal execution, artifacts, result submission.

### M4 Economics
Implement JouleWork, quality scoring, reliability snapshots, pricing, ledger, reserves.

### M5 UI
Implement Svelte dashboards for tasks, approvals, fleet, ledger, reliability, settings.

### M6 Integrations
Implement Paperclip connector and projected workflow state.

### M7 Advanced workloads
Implement code patch merge flows and autosearch self-improve bounded loops.

## Non-negotiable rules

1. Keep multi-tenancy.
2. Use Svelte, not React.
3. Keep Paperclip as required dependency/integration target.
4. Keep Hermes as only production runtime in v1.
5. Postgres is durable truth; NATS is realtime only.
6. Keep approval-first design.
7. Keep the ledger model.
8. Keep isolated ephemeral marketplace workspaces by default.
9. Do not store large artifacts in Postgres.
10. Implement clean service/package boundaries.

## First output

Generate:
- repo tree
- root config files
- migration plan
- domain package
- docker-compose
- package skeletons
- TODO map for milestones

Then implement milestone 1 completely.

## Preserve external anchors
Paperclip:
- https://paperclip.ing/
- https://github.com/paperclipai/paperclip

Hermes:
- https://hermes-agent.nousresearch.com/
- https://hermes-agent.nousresearch.com/docs/getting-started/installation/
- https://hermes-agent.nousresearch.com/docs/user-guide/features/tools/
- https://hermes-agent.nousresearch.com/docs/user-guide/security/
- https://hermes-agent.nousresearch.com/docs/user-guide/features/mcp/

SporeMesh inspiration:
- https://www.sporemesh.com/
- https://www.piwheels.org/project/sporemesh

Karpathy context:
- https://github.com/karpathy/nanochat
- https://github.com/karpathy/nanoGPT
- https://github.com/karpathy/autosearch
