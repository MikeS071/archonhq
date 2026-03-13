# SEQUENCE_DIAGRAMS.md

## Happy path: task to settlement
Operator UI -> API: create task
API -> DB: task + approval request + task.created + approval.requested
Approver UI -> API: approve
API -> DB: approval.approved
Scheduler -> DB: lease.granted
Node -> API: claim lease
Node -> Hermes: execute
Node -> Object Store: upload artifacts
Node -> API: submit result
Verifier -> API/DB: verification.completed
Reducer -> DB: reduction.accepted
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
