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
- acceptance_contract.*
- verification.*
- validation.*
- reduction.*
- reliability.*
- market.*
- escrow.*
- payout.*
- dispute.*
- simulation.*
- pricing.*
- ledger.*
- policy.*
- notification.*

## Sample progression
task.created
task.acceptance_contract_attached
approval.requested
approval.approved
lease.granted
result.submitted
verification.completed
validation.run_started
validation.stage_passed
validation.stage_rejected
validation.run_completed
reduction.accepted
ledger.settlement_posted
ledger.reserve_hold_created
market.listing_created
market.claim_awarded
escrow.locked
payout.requested
dispute.opened
simulation.run_started
simulation.finding_created
simulation.run_completed
