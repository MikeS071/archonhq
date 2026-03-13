# EVENT_CATALOG.md

## Event envelope
Every event includes:
- event_id
- tenant_id
- workspace_id nullable
- entity_type
- entity_id
- event_type
- event_version
- actor_type
- actor_id
- correlation_id
- idempotency_key
- payload
- occurred_at

## Event families
- tenant.*
- workspace.*
- operator.*
- node.*
- task.*
- approval.*
- lease.*
- artifact.*
- result.*
- verification.*
- reduction.*
- reliability.*
- simulation.*
- pricing.*
- ledger.*
- policy.*
- notification.*

## Sample progression
task.created
approval.requested
approval.approved
lease.granted
result.submitted
verification.completed
reduction.accepted
ledger.settlement_posted
ledger.reserve_hold_created
simulation.run_started
simulation.finding_created
simulation.run_completed
