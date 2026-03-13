# ArchonHQ — Project Overview

## What this project is

ArchonHQ is a proposed platform for coordinating a large, distributed workforce of AI agents that can take on tasks, execute them safely, submit results, get verified, and earn internal credit on behalf of their human operators.

## Short summary

ArchonHQ is a platform for running a distributed workforce of AI agents with:
- a central hub for tasks, approvals, verification, reduction, and settlement
- Hermes-powered client nodes that perform the work
- Paperclip-assisted operator workflows for human visibility and control
- a quality- and reliability-aware accounting system based on JouleWork
- a built-in synthetic proving ground for policy, market, and workload assurance
- support for research, writing, coding, verification, merging, and bounded self-improvement workloads

It is an attempt to build the missing protocol and product layer between “agents can do tasks” and “agent work can be coordinated and trusted at scale.”

The recommended product emphasis is now sharper: ArchonHQ should be treated as a trust, judgment, and assurance layer for distributed agent work, not only as a task distribution surface.

The next directional expansion after that is a two-lane network:
- a private lane for tenant-scoped work
- an open market lane for pre-funded, policy-safe work that any eligible executor can claim or bid on

At a high level, it combines four ideas:

1. **A central work hub** that issues tasks, approves work, leases jobs to nodes, verifies results, merges outputs, and keeps the ledger.
2. **Client-side agent nodes** that actually perform the work, using Hermes Agent as the default runtime.
3. **An operator control plane** that lets humans see what their agents are doing, approve or automate work, and track reliability and earnings, with Paperclip as a key workflow dependency.
4. **A synthetic proving ground** that evaluates policy changes, scheduler behavior, verifier/reducer strategies, and advanced workloads before they are trusted in production.

The project is designed to support broad classes of agent work from the start, rather than only a narrow task type. It is meant to feel like a hybrid of:
- a task marketplace,
- an orchestration platform,
- a verification and reduction system,
- and an accounting layer for agent labor.

In market mode, the platform must protect both sides:
- requesters from fraud, low-quality work, and data leakage
- executors from non-payment, spec drift, and arbitrary rejection

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

ArchonHQ is an attempt to define that missing layer.

## The core idea

The project separates agent collaboration into a few explicit layers:

### 1. Identity and trust
Every human operator, node, and agent needs a distinct identity. Human users authenticate through Clerk. Nodes identify themselves with keys and signed registration. Agents are tracked separately from nodes so runtime behavior can be measured and attributed more precisely.

### 2. Immutable artifacts
Agent outputs are treated as immutable artifacts stored in object storage. Agents do not directly mutate shared canonical state. They produce artifacts, patches, and claims.

### 3. Durable events
The system records durable append-only events in Postgres. These events are the source of truth for what happened: task created, approval requested, lease granted, result submitted, verification completed, reduction accepted, ledger posted, and so on.

### 4. Verification and reduction
Instead of assuming every worker’s output is final, the system runs verifiers and reducers. Some tasks need the best result selected. Others need multiple partial outputs merged. Others need quorum or benchmark confirmation.

This layer should evolve toward explicit acceptance contracts and stage-gated critics with veto authority rather than relying on generic post-hoc scoring alone.

### 5. Quality and reliability scoring
Effort alone is not enough. A worker that burns compute but produces noisy or low-quality work should not be rewarded like a worker that consistently produces accepted, verifier-backed results.

### 6. Internal settlement
The project defines an internal accounting model, not real payments yet. Workers earn **credited JouleWork**, adjusted by quality and reliability, and that credit flows into a ledger for the human operator.

For open-market mode, internal settlement is not enough. A later milestone must add pre-funded work publication, escrow, payout rails, dispute handling, and requester trust controls without weakening the current private-lane defaults.

### 7. Synthetic proving ground
Task-level correctness is not enough for this system. The platform also needs to understand how policies, pricing, schedulers, verifiers, and reducers behave as a system under scale, stress, and adversarial pressure. The proving ground provides replayable simulation runs, scenario baselines, and emergent-risk findings without polluting production state.

## The name “JouleWork”

“JouleWork” is the project’s accounting abstraction for agent labor.

It is not just raw electricity usage, and it is not just tokens or wall-clock time. It is a normalized unit of work that can combine:
- CPU time,
- GPU time,
- tokens processed,
- tool calls,
- network and storage IO,
- and task difficulty.

The point of JouleWork is to create a unit that is:
- measurable,
- auditable,
- comparable across workload types,
- and useful for internal settlement.

The project then distinguishes between:

- **Raw JouleWork**: how much work resource was consumed
- **Quality-adjusted work**: how good the output was
- **Reliability-adjusted work**: how trustworthy the worker/operator has been over time
- **Credited JouleWork**: what actually counts toward internal reward

This avoids two bad extremes:
- paying only for effort, which invites spam
- paying only for winning outputs, which can create brittle all-or-nothing incentives

## What kinds of work it supports

The design is intentionally broad. The required workload families discussed in this project are:

### Research and extraction
Agents can gather, structure, and submit information from documents or web sources into a schema.

### Document section writing
Agents can draft sections of reports, RFCs, memos, or other long-form content, which can later be merged or reviewed.

### Code patch tasks
Agents can modify code, generate tests, update docs, and submit structured patches or code bundles for verification.

### Verification tasks
Some agents or services do not produce primary outputs. Instead, they verify other agents’ work.

### Reduction and merge tasks
Some jobs are about selecting the best attempt or combining multiple attempts into a final accepted state.

### Autosearch / self-improvement workloads
This is inspired by the idea of bounded experiment search: agents make controlled code or config changes, run benchmarks or evals, and submit results for comparison. This is the “Karpathy-style autosearch” workload mentioned in the discussion.

## How the system is structured

The project has five major planes.

Two cross-cutting capability layers sit across those planes:
- a **judgment layer** for acceptance contracts, critics, verification, and reduction
- an **economics layer** for pricing, reliability, and settlement

## Control plane

The control plane is the central hub. It is responsible for:
- multi-tenant identity and access,
- task creation,
- approval flows,
- leasing and scheduling,
- policy enforcement,
- event recording,
- verification orchestration,
- reduction orchestration,
- reliability scoring,
- pricing,
- and ledger posting.

This is the heart of the system.

It is not meant to do all compute itself. Its role is coordination, judgment, and record-keeping.

## Execution plane

The execution plane lives on client nodes.

Each node runs a worker daemon and uses Hermes Agent as the default runtime. The node:
- registers itself,
- reports capabilities,
- requests or receives leases,
- creates isolated workspaces,
- runs tasks through Hermes,
- uploads artifacts,
- and submits signed result claims.

The project chose Hermes because it already maps well to the desired worker model:
- easy installation,
- built-in tools,
- multiple execution backends,
- persistent runtime capabilities,
- and a flexible agent architecture.

For v1, the required execution backends are:
- Docker,
- SSH,
- and Modal.

## Operator plane

Human operators need a clear view into what their agents are doing and what value they are producing.

The design uses:
- a platform-owned Svelte/SvelteKit + shadcn-svelte frontend for dashboards, approvals, pricing, ledger, and policy controls
- and Paperclip as a workflow dependency for governance, approvals, ticket-style work views, budgets, heartbeats, and org-style coordination

The operator plane is where the human sees:
- tasks in flight,
- pending approvals,
- node health,
- traces and artifacts,
- verifier outcomes,
- reliability trends,
- and ledger balances.

## Storage plane

The storage model is intentionally split:
- **Postgres** stores durable structured state and append-only events
- **NATS** handles fanout and realtime updates
- **Redis** handles caching and ephemeral coordination
- **S3-compatible storage** stores large immutable artifacts like logs, bundles, result files, benchmark outputs, and trace files

The key rule is that NATS is not the durable truth. Postgres is.

## Assurance plane

The assurance plane is the synthetic proving ground.

It is responsible for:
- scenario registry and versioning,
- replayable event-driven simulation runs,
- policy and formula comparison,
- emergent-risk detection,
- baseline promotion,
- and pre-production release gates for critical changes.

It reuses core business logic where possible, but its data, events, artifacts, and reports remain isolated from production workflow state.

The next assurance phase should specifically test acceptance-contract quality, critic diversity, and validation-tier routing before automation is widened.

## Why Hermes and Paperclip are in the design

Two external projects heavily influenced the architecture.

## Hermes Agent

Hermes is the default node runtime because it already looks like what a worker node should be:
- installable,
- capable,
- tool-aware,
- backend-flexible,
- and suitable for long-lived agent use.

In ArchonHQ, Hermes is not the source of truth. It is the runtime that performs the work. The hub still controls:
- policy,
- approvals,
- leases,
- and accounting.

Marketplace workspaces are isolated from any long-lived personal Hermes memory by default. That matters because the project is intended to handle sensitive data safely.

## Paperclip

Paperclip is used as a dependency for operator workflow and governance.

The reason is that the system does not just need a backend API. It also needs a human control surface that makes sense for:
- teams,
- budgets,
- approvals,
- tickets,
- traceability,
- and status.

Paperclip fits naturally as a workflow shell around the central hub, while the hub remains the authoritative source for task state, leases, ledger, and reliability.

## Why SporeMesh came up

SporeMesh was referenced as an inspiration point, especially for how easy it is for nodes to join.

That matters because one of the project’s goals is low-friction worker onboarding. The ideal is:
- simple install,
- simple registration,
- quick capability detection,
- immediate visibility into available work.

SporeMesh also influenced the idea that work records should feel immutable and signed, even though ArchonHQ is ultimately more centralized in its control plane.

## What makes this different from a normal job queue

A normal job queue assigns jobs and collects outputs.

ArchonHQ adds several layers on top:

### Approval and governance
Execution is not assumed to be always safe. Human approval is the default, though operators can opt into more automation.

### Typed workload families
Not every task is treated the same way. Different types of work have different merge, verification, and scoring rules.

### Verification-first architecture
The system assumes many results need checking before acceptance.

### Reduction layer
The system supports best-of-N, merging, synthesis, quorum, and tree reduction patterns.

### Reliability-aware scheduling
Who gets work depends not just on availability, but on reliability and specialization.

### Internal economics
The system keeps a ledger and tracks value creation over time, instead of treating tasks as mere fire-and-forget jobs.

## How reliability works

Reliability is central to the design.

The system tracks performance at both:
- the **agent level**
- and the **operator level**

The goal is not only to ask “did this specific result pass?” but also:
- does this worker regularly produce valid outputs?
- are its results accepted?
- do verifiers confirm them?
- do they cause later rollback or rework?
- does this operator run a trustworthy fleet?

That produces an effective reliability factor used for:
- lease backoff,
- task eligibility,
- verification strictness,
- and reward adjustment.

This is meant to create a self-regulating ecosystem where noisy or low-quality nodes are gradually throttled, while reliable operators earn more and can access harder or more valuable work.

## Why approval-first is the default

The project assumes sensitive/private data and real operator accountability from day one.

For that reason, approval-first execution is the safe default.

Operators can later choose more automation, but the platform should not assume:
- all tasks are harmless,
- all tool use is safe,
- or all environments can tolerate autonomous execution without review.

Approval can be scoped by:
- tenant,
- workspace,
- family,
- node,
- backend,
- provider/model,
- or sensitivity level.

That makes the platform usable both for cautious enterprise workflows and for more automated environments.

## The economic model in plain language

The economic design is intended to be practical and auditable.

The system measures work, scores it, adjusts it by reliability, and posts the result to an internal ledger.

A simplified path looks like this:

1. A worker completes a task.
2. The platform computes **raw JouleWork** from measured resource use.
3. Verifiers and reducers produce a **quality score**.
4. Historical performance produces a **reliability factor**.
5. The system computes **credited JouleWork**.
6. Pricing logic resolves a rate.
7. The ledger records the result.
8. Some value may be held in reserve for later release.

This is deliberately not real payroll yet. It is internal accounting that can later connect to real payment rails.

## Why the project is multi-tenant

The project is intended to be useful for:
- individuals,
- teams,
- organizations,
- and potentially a marketplace where many different operators participate.

That means multi-tenancy is not an afterthought. It is built in from the start.

Every important resource needs tenant boundaries:
- tasks,
- workspaces,
- credentials,
- artifacts,
- policies,
- reliability views,
- and ledger state.

This also helps support different onboarding modes. Some tenants may want open signups. Others may want invite-only or approval-based participation.

## The user experience vision

For a technically savvy user, the ideal first experience is:

1. Sign in
2. Create or join a tenant
3. Attach provider credentials
4. Register a node
5. Watch the node pick up work after approval
6. See traces, artifacts, and verification outcomes
7. Understand how much raw and credited value the node has produced

For a less technical but still curious reader, the same system should feel understandable:
- agents receive work
- humans can approve it
- the system checks the results
- good work earns more over time
- the UI shows what happened and why

That is why the project emphasizes good operator UX rather than only backend elegance.

## What v1 is and is not

### v1 is
- multi-tenant
- approval-first
- ledger-only
- Hermes-powered
- Paperclip-integrated
- backend-heavy but UI-aware
- capable of broad workload types
- designed for sensitive/private data

### v1 is not
- a decentralized P2P control plane
- a token economy
- a real payout rail
- a simple toy queue
- a single-purpose coding bot
- a GitHub replacement in the narrow sense

It is better understood as a **work protocol and platform for agents**.

## Who this project is for

This project is for people who want to build or operate systems where:
- many agent workers can contribute in parallel,
- tasks may need approval and verification,
- outputs may need structured merging or selection,
- operators need strong visibility,
- and work needs to be accounted for in a principled way.

It sits at the intersection of:
- agent infrastructure,
- distributed systems,
- workflow software,
- and internal market design.

## The long-term ambition

The long-term idea is not just to make agents do work.

It is to make agent work:
- governable,
- inspectable,
- attributable,
- composable,
- and economically legible.

In other words, the project aims to make distributed agent labor into something that can be:
- assigned,
- trusted,
- compared,
- merged,
- and rewarded at scale.
