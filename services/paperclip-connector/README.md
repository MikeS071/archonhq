# paperclip-connector

Service boundary for projecting internal ArchonHQ read models into Paperclip workflow surfaces.

## Responsibilities

- Build tenant-scoped projection payloads from Postgres read models.
- Sync workspace summaries, approval queue state, fleet heartbeat summaries, settlement snapshots, and reliability snapshots.
- Enforce `source_of_truth=postgres` in all outbound payloads.
- Return sync metadata for API status tracking.

## Guardrail

Paperclip is a projection target only and is never authoritative for task, lease, ledger, reliability, or event truth.
