# simulation

Service scaffold for synthetic proving-ground workflows, scenario orchestration, replayable runs, and emergent-risk reporting.

## Boundary

- Owns simulation scenario registry and run orchestration.
- Produces simulation findings, baseline comparisons, and replay metadata.
- Reuses shared policy/scoring/verifier/reducer packages.
- Must not write into production task/lease/result/ledger/reliability tables.

## Planned Package Layout

```text
services/simulation/
├─ cmd/simulation/main.go
├─ internal/
│  ├─ api/
│  │  ├─ handlers.go
│  │  └─ dto.go
│  ├─ service/
│  │  ├─ scenario_service.go
│  │  ├─ run_service.go
│  │  ├─ baseline_service.go
│  │  └─ replay_service.go
│  ├─ engine/
│  │  ├─ deterministic_stub.go
│  │  ├─ sampled_synthetic.go
│  │  └─ runtime_backed.go
│  ├─ findings/
│  │  └─ rules.go
│  ├─ repository/
│  │  ├─ scenario_repo.go
│  │  ├─ run_repo.go
│  │  ├─ baseline_repo.go
│  │  └─ replay_repo.go
│  ├─ projections/
│  │  └─ consumers.go
│  └─ model/
│     └─ types.go
└─ README.md
```

## Phase 1 Deliverables

- [x] Deterministic stub mode only.
- [x] Scenario create/list/get + version + publish APIs.
- [x] Run create/list/get + cancel APIs.
- [x] Run events/metrics/findings/artifacts read APIs.
- [x] Baseline promote/list/get + compare API.
- [x] Replay request/get API with approval-gated path for sensitive sources.
