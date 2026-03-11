# CODEX_INITIAL_PROMPT.md

Build a new repository named **joulework-network**.

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
