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
