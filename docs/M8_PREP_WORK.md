# M8 Prep Work Plan

## Scope

Prep work translates the simulation specification into implementation-ready artifacts without shipping runtime behavior yet.

## Completed Prep Artifacts

- Simulation API contract reflected in [docs/openapi/openapi.yaml](./openapi/openapi.yaml)
- Simulation migration breakdown in [docs/M8_MIGRATION_SPEC.md](./M8_MIGRATION_SPEC.md)
- Simulation service package layout defined in [services/simulation/README.md](../services/simulation/README.md)

## Execution Order (Implementation)

1. Finish verifier/reducer production APIs and remove placeholder handlers.
2. Implement `services/simulation` with deterministic stub mode.
3. Apply simulation schema + read-model migrations.
4. Add scenario library seeding + baseline promotion + compare flow.
5. Enforce simulation-gated rollout checks for critical policy categories.

## Concrete Task Breakdown

### A. API and Contracts

- Add `/v1/simulation/*` handlers in `apps/api/internal/httpserver`.
- Reuse existing error envelope and correlation-id/idempotency middleware.
- Add request validation for scenario definitions and run start constraints.
- Wire replay endpoints to elevated approval checks for sensitive sources.

### B. Service Scaffolding

- Create simulation service layer and repositories under `services/simulation/internal`.
- Add deterministic stub engine with seeded reproducibility.
- Define findings rules and rollout verdict generation interface.

### C. Data and Events

- Implement migrations from `docs/M8_MIGRATION_SPEC.md`.
- Emit required `simulation.*` events.
- Add consumers/materializers for simulation read models only.

### D. UI and Gates

- Add simulation pages in web app for scenarios, runs, findings, and comparisons.
- Add release gate checks for verifier/reducer/scheduler/pricing/reliability policy changes.
- Surface machine-readable compare verdicts in approval/release workflows.

## Definition of Prep Complete

- OpenAPI includes simulation routes, schemas, and mutation idempotency requirements.
- Migration work is partitioned into concrete files with ordering and rollback posture.
- Service boundary/layout is documented with responsibilities and phase-1 deliverables.
