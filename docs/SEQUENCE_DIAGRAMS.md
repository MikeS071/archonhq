# SEQUENCE_DIAGRAMS.md

## Happy path: task to settlement
Operator UI -> API: create task
API -> DB: task + acceptance contract snapshot + approval request + task.created + task.acceptance_contract_attached + approval.requested
Approver UI -> API: approve
API -> DB: approval.approved
Validation planner -> DB: validation.run_started (plan gate)
Plan critic -> DB: validation.stage_passed
Scheduler -> DB: lease.granted
Node -> API: claim lease
Node -> Hermes: execute
Node -> Object Store: upload artifacts
Node -> API: submit result
Execution critic -> DB: validation.stage_passed or validation.stage_rejected
Verifier -> API/DB: verification.completed + validation.stage_passed
Reducer -> DB: reduction.accepted
Output/policy critic -> DB: validation.run_completed
Ledger -> DB: ledger.settlement_posted + reserve hold
Projections -> UI: refresh

## Happy path: simulation policy gate
Operator UI -> API: create simulation scenario version
API -> DB: simulation.scenario_version_created
Operator UI -> API: start simulation run with candidate policy
Simulation -> DB: simulation.run_started
Simulation -> NATS/DB: simulation.step_recorded + simulation.metric_recorded
Findings engine -> DB: simulation.finding_created
Simulation -> DB: simulation.run_completed
Reviewer UI -> API: compare against baseline
Reviewer -> API: promote or reject candidate policy

## Happy path: validation escalation
Critic -> API/DB: validation.stage_needs_review
Operator UI -> API: inspect validation run and evidence refs
Approver UI -> API: validation escalation decision
API -> DB: validation.escalated + approval.approved or approval.denied

## Happy path: open-market listing to payout
Requester -> API: create market listing
API -> DB: task + market contract snapshot + market.listing_created
Requester -> API: fund listing
API -> DB: escrow.funded
Executor -> API: create claim or bid
API -> DB: market.claim_created or market.bid_submitted
Requester/Matcher -> API: award claim or accept bid
API -> DB: market.claim_awarded + escrow.locked
Executor Node -> Hermes: execute
Executor -> API: submit result
Validation/Reduction -> DB: accepted output
API -> DB: payout.requested + escrow.released
Payout service -> rail: transfer funds
Payout service -> DB: payout.completed

## Happy path: open-market dispute
Requester or Executor -> API: open dispute
API -> DB: dispute.opened
Arbitration critics -> API/DB: decision bundle
Arbitrator -> API: resolve dispute
API -> DB: dispute.resolved + escrow.released or escrow.refunded + payout.completed or payout.failed
