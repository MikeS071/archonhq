# JouleWork Network — Complete Codex Build Kit

This is the merged, reconciled handoff package for building **JouleWork Network**.

It combines all prior specs in this chat into one GitHub-ready kit.

## What this kit contains

Core:
- `README.md`
- `SPEC.md`
- `MONOREPO_STRUCTURE.md`
- `CODEX_INITIAL_PROMPT.md`

Architecture and contracts:
- `docs/API_CONTRACTS.md`
- `docs/DB_SCHEMA_DRAFT.md`
- `docs/EVENT_CATALOG.md`
- `docs/NATS_SUBJECT_MAP.md`
- `docs/UI_WIREFRAMES.md`
- `docs/SEQUENCE_DIAGRAMS.md`
- `docs/ADRS.md`
- `docs/CODEX_MILESTONE_PROMPTS.md`
- `docs/ERROR_MODEL.md`
- `docs/SECURITY_MODEL.md`
- `docs/POLICY_SCHEMA.md`
- `docs/HERMES_ADAPTER_SPEC.md`
- `docs/PAPERCLIP_CONNECTOR_SPEC.md`
- `docs/OBSERVABILITY_SPEC.md`
- `docs/TEST_PLAN.md`

Machine-readable starters:
- `docs/openapi/openapi.yaml`
- `docs/schemas/*.json`

Code/reference starters:
- `docs/go_interfaces/interfaces.go`
- `examples/go/*.go`
- `examples/sql/*.sql`
- `examples/json/*.json`
- `examples/api/*.md`

Infra/deploy starters:
- `deploy/templates/docker-compose.yml`
- `deploy/templates/.env.example`

Database:
- `migrations/*.sql`

Frontend:
- `frontend/FRONTEND_ROUTE_COMPONENT_MAP.md`

## Important honesty note

This is the most complete build kit I can generate here, but some parts are still **first-pass implementation specifications**, not legally or operationally certified production artifacts. In particular:
- compliance is still “global best-effort”
- OpenAPI is comprehensive starter-level, not exhaustively enumerated for every variant
- SQL schema is detailed and coherent, but still needs engineering validation during implementation
- Paperclip and Hermes integration details are specified from their public product/docs posture and the design choices made in this chat

## Fixed product decisions from this chat

- onboarding supports both open and permissioned modes
- v1 is ledger-only, no external payouts yet
- bring-your-own-key inference only in v1
- multi-tenant from day one
- broad workload support from day one
- Paperclip is a required dependency in v1
- Hermes is the only production runtime in v1
- Docker, SSH, and Modal are required execution backends in v1
- frontend is Svelte/SvelteKit
- backend is Go
- durable truth is Postgres; NATS is realtime fanout
- auth uses Clerk
- private/sensitive data is supported by default
- marketplace workspaces are isolated and ephemeral by default
- human approval is default; automation is optional
- pricing supports both fixed rates and dynamic bidding
- optimize for long-term architecture and good UX

## External project anchors discussed in this chat

Paperclip:
- https://paperclip.ing/
- https://github.com/paperclipai/paperclip

Hermes Agent:
- https://hermes-agent.nousresearch.com/
- https://hermes-agent.nousresearch.com/docs/getting-started/installation/
- https://hermes-agent.nousresearch.com/docs/user-guide/features/tools/
- https://hermes-agent.nousresearch.com/docs/user-guide/security/
- https://hermes-agent.nousresearch.com/docs/user-guide/configuration/
- https://hermes-agent.nousresearch.com/docs/user-guide/features/mcp/

SporeMesh inspiration:
- https://www.sporemesh.com/
- https://www.piwheels.org/project/sporemesh
- https://data.safetycli.com/packages/pypi/sporemesh/

Karpathy-related context:
- https://github.com/karpathy/nanochat
- https://github.com/karpathy/nanoGPT
- https://github.com/karpathy/autosearch

## How to use this with Codex

1. Create a new GitHub repo, e.g. `joulework-network`
2. Unzip this kit into the repo root
3. Commit the kit as the initial commit
4. Load `SPEC.md`, `CODEX_INITIAL_PROMPT.md`, and `docs/CODEX_MILESTONE_PROMPTS.md` into Codex
5. Build milestone by milestone
