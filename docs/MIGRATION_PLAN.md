# Migration Plan

## Goals

- Preserve `migrations/000_full_schema.sql` as baseline schema reference.
- Move to incremental, ordered SQL migrations for forward-only delivery.
- Keep Postgres as durable truth and avoid large artifact bytes in DB.

## Plan

1. Baseline
- Keep `000_full_schema.sql` untouched as canonical bootstrap reference.
- Add migration naming convention: `NNN_description.up.sql` and `NNN_description.down.sql`.

2. Incremental phases
- M1: bootstrap migration framework and metadata table for applied versions.
- M2: add indexes/constraints for workflow paths and projection read models.
- M3: add runtime execution, artifact metadata, and signature verification constraints.
- M4: add economics tables refinements for pricing/ledger/reserves.
- M5+: add UI support read models and integration tracking tables as needed.

3. Data safety rules
- All migration scripts must be idempotent for retries where possible.
- Backfills run separately from schema changes for large tables.
- Lock-heavy operations use off-peak deployment windows.

4. Rollback posture
- Prefer forward-fix; use `.down.sql` for local/dev recovery.
- High-risk migrations require pre/post checks and backup verification.

## First Planned Migration Files

- `migrations/001_migration_metadata.up.sql`
- `migrations/001_migration_metadata.down.sql`
- `migrations/002_projection_read_models.up.sql`
- `migrations/002_projection_read_models.down.sql`
