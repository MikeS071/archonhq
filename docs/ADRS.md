# ADRS.md

## ADR-001 Hybrid event architecture
Postgres is durable truth; NATS is fanout/realtime.

## ADR-002 Hermes-only v1 runtime
Use Hermes only in production for v1 to reduce integration surface.

## ADR-003 Paperclip as dependency not source of truth
Paperclip is used for operator workflow and governance projection only.

## ADR-004 Approval-first default
Sensitive/private work and trust concerns require human approval by default.

## ADR-005 Ledger-only v1
Keep internal accounting now, leave payout rails abstracted for later.

## ADR-006 Built-in synthetic proving ground
Critical policy, verifier, reducer, scheduler, and economics changes require isolated simulation and replay before broader production rollout.

## ADR-007 Acceptance-contract and critic-gated trust model
Tasks requiring trusted acceptance must declare success criteria before execution and advance through stage-gated critics with veto authority rather than producer self-certification.

## ADR-008 Two-lane network model
ArchonHQ keeps a private tenant lane as the default posture and adds open-market mode only as a separate lane with funded listings, escrow, payout, dispute, and work-class controls.
