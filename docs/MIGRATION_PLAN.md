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
- M8: add acceptance-contract, validation, and simulation tables, read models, and isolation constraints.
  - Detailed migration breakdown: `docs/M8_MIGRATION_SPEC.md`.
- M9: add market-mode profiles, listings, escrow, payouts, disputes, and market read models.
  - Detailed migration breakdown: `docs/M9_MIGRATION_SPEC.md`.

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

## Planned M8 Migration Files

- `migrations/005_acceptance_contracts_core.up.sql`
- `migrations/005_acceptance_contracts_core.down.sql`
- `migrations/006_validation_pipeline_core.up.sql`
- `migrations/006_validation_pipeline_core.down.sql`
- `migrations/007_validation_indexes_and_read_models.up.sql`
- `migrations/007_validation_indexes_and_read_models.down.sql`
- `migrations/008_simulation_schema_core.up.sql`
- `migrations/008_simulation_schema_core.down.sql`
- `migrations/009_simulation_schema_runtime.up.sql`
- `migrations/009_simulation_schema_runtime.down.sql`
- `migrations/010_simulation_indexes_and_isolation.up.sql`
- `migrations/010_simulation_indexes_and_isolation.down.sql`
- `migrations/011_simulation_read_models.up.sql`
- `migrations/011_simulation_read_models.down.sql`

## Planned M9 Migration Files

- `migrations/012_market_profiles_and_listings_core.up.sql`
- `migrations/012_market_profiles_and_listings_core.down.sql`
- `migrations/013_market_claims_bids_and_reputation.up.sql`
- `migrations/013_market_claims_bids_and_reputation.down.sql`
- `migrations/014_escrow_and_payout_core.up.sql`
- `migrations/014_escrow_and_payout_core.down.sql`
- `migrations/015_market_disputes_and_read_models.up.sql`
- `migrations/015_market_disputes_and_read_models.down.sql`
