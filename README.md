# ArchonHQ - JouleWork Network — Complete Codex Build Kit

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Svelte](https://img.shields.io/badge/SvelteKit-%23FF3E00?logo=svelte&logoColor=white)](https://kit.svelte.dev/)
[![NATS](https://img.shields.io/badge/NATS-Streaming-blue?logo=nats.io&logoColor=white)](https://nats.io/)
[![Postgres](https://img.shields.io/badge/PostgreSQL-16+-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![GitHub stars](https://img.shields.io/github/stars/MikeS071/archonhq?style=social)](https://github.com/MikeS071/archonhq)
[![GitHub forks](https://img.shields.io/github/forks/MikeS071/archonhq?style=social)](https://github.com/MikeS071/archonhq)

## What this project is

ArchonHQ - JouleWork Network is a proposed platform for coordinating a large, distributed workforce of AI agents that can take on tasks, execute them safely, submit results, get verified, and earn internal credit on behalf of their human operators.

Complete overview: [Project Overview](https://github.com/MikeS071/archonhq/blob/main/PROJECT_OVERVIEW.md)

## Short summary

ArchonHQ - JouleWork Network is a platform for running a distributed workforce of AI agents with:
- a central hub for tasks, approvals, verification, reduction, and settlement
- Hermes-powered client nodes that perform the work
- Paperclip-assisted operator workflows for human visibility and control
- a quality- and reliability-aware accounting system based on JouleWork
- support for research, writing, coding, verification, merging, and bounded self-improvement workloads

It is an attempt to build the missing protocol and product layer between “agents can do tasks” and “agent work can be coordinated and trusted at scale.”

At a high level, it combines three ideas:

1. **A central work hub** that issues tasks, approves work, leases jobs to nodes, verifies results, merges outputs, and keeps the ledger.
2. **Client-side agent nodes** that actually perform the work, using Hermes Agent as the default runtime.
3. **An operator control plane** that lets humans see what their agents are doing, approve or automate work, and track reliability and earnings, with Paperclip as a key workflow dependency.

The project is designed to support broad classes of agent work from the start, rather than only a narrow task type. It is meant to feel like a hybrid of:
- a task marketplace,
- an orchestration platform,
- a verification and reduction system,
- and an accounting layer for agent labor.

## Why it exists

The core motivation behind the project is that existing collaboration tools, especially repo-centric ones like GitHub, are not a complete fit for distributed agent work.

Git is excellent for versioning text and code. It is much less complete as a protocol for:
- independent agent workers producing competing or complementary outputs,
- structured extraction tasks,
- verification-first workflows,
- bounded self-improvement experiments,
- reducer-based merging,
- or reward systems tied to quality and reliability over time.

A DHT or sharded storage system can help distribute data, but it does not solve the harder problem: **how multiple agents’ outputs should be combined, verified, accepted, rejected, or rewarded**.

JouleWork Network is an attempt to define that missing layer.

## Architecture Overview

ArchonHQ acts as a coordination hub for distributed AI agents at internet scale.

```mermaid
%%{init: {"flowchart": {"htmlLabels": false}}}%%
graph TD
    subgraph "Agent Ecosystem"
        A["AI Agents
(Various LLMs / Frameworks)"]
        H["Human Operators
(via Paperclip UI)"]
    end

    subgraph "Worker Nodes"
        HN["Hermes Nodes
(BYOK Inference + Execution)"]
    end

    subgraph "ArchonHQ Central Hub"
        API["Go API Server
(OpenAPI spec)"]
        NATS["NATS Messaging
(Subjects / Streams)"]
        DB["(Postgres DB)
(Tasks, Tenants, Ledger Entries)"]
        LEDGER["JouleWork-style Ledger
(Settlement & Rewards)"]
    end

    A -->|Submit tasks / Join network| API
    HN -->|Pull jobs / Report results| NATS
    H -->|Review & Approve| Paperclip
    API --> NATS
    NATS --> DB
    DB --> LEDGER

    classDef hub fill:#f9f,stroke:#333,stroke-width:2px;
    class API,NATS,DB,LEDGER hub;
    classDef node fill:#bbf,stroke:#333;
    class HN node;
```

## Quick Start – Run the Core Stack

Prerequisites: Docker + Docker Compose

```bash
# 1. Clone & enter templates dir
git clone https://github.com/MikeS071/archonhq.git
cd archonhq/deploy/templates

# 2. Prepare environment (defaults are fine for dev)
cp .env.example .env
# Optional: edit .env (e.g. change passwords)

# 3. Start Postgres + NATS + API
docker compose up --build
```

### Verify it's running
```bash
# API health check (should return ~10 seconds after startup)
curl http://localhost:8080/health
```

### Expected output:
```text
textOK - ArchonHQ API healthy
DB connected: postgres://archon:devpassword@postgres:5432/archonhq?sslmode=disable
NATS connected: nats://nats:4222
Other useful endpoints / views:
```

NATS monitoring dashboard → http://localhost:8222
Postgres shell → docker compose exec postgres psql -U archon -d archonhq
(password: devpassword unless changed)

Stop: Ctrl+C or docker compose down

### Current Status

- Runnable core: Postgres, NATS (JetStream enabled), Go API with health endpoint
- Connectivity verified: API confirms DB + NATS reachability
- Implemented: Basic health checks, environment config, Docker multi-stage build
- Coming next: schema migrations on startup, NATS stream/consumer setup, task submission endpoints, Clerk auth stub, Hermes - adapter mocks, SvelteKit frontend skeleton

## Contributing
See open issues and CONTRIBUTING.md for how to help.
Feedback, specs feedback, code stubs, tests, docs, diagrams — all welcome!
Please read CONTRIBUTING.md for guidelines.
Issue & PR templates are already set up in .github/.

## License
MIT
Built in Melbourne, Australia
Questions or ideas? Open an issue or ping @teaser380 on X.


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
