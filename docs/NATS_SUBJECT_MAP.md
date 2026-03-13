# NATS_SUBJECT_MAP.md

## Convention
`jw.{tenant_id}.{family}.{event}`

## Examples
- jw.ten_01.task.created
- jw.ten_01.approval.requested
- jw.ten_01.lease.granted
- jw.ten_01.result.submitted
- jw.ten_01.verification.completed
- jw.ten_01.reduction.accepted
- jw.ten_01.ledger.settlement_posted
- jw.ten_01.simulation.run_started
- jw.ten_01.simulation.finding_created
- jw.ten_01.simulation.run_completed

## Durable consumer groups
- projections
- scheduler
- approvals
- verification
- reductions
- reliability
- simulation
- simulation-projections
- ledger
- web-realtime
- paperclip-connector

Simulation consumers must not materialize into production workflow read models.
