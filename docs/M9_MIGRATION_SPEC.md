# M9 Migration Spec (Open Market Network)

## Purpose

Define the migration units required to introduce open-market mode after the M8 judgment and assurance foundation is in place.

This migration spec assumes:

- acceptance contracts
- validation runs
- simulation baselines

already exist.

## Migration Units

### 012 `market_profiles_and_listings_core`

Files:
- `migrations/012_market_profiles_and_listings_core.up.sql`
- `migrations/012_market_profiles_and_listings_core.down.sql`

Creates:
- `market_profiles`
- `market_profile_verifications`
- `market_listings`
- `market_listing_shards`

Core constraints:
- `profile_type` check (`requester`, `executor`, `hybrid`)
- `work_class` check (`public_open`, `public_sealed`, `restricted_market`, `private_tenant_only`)
- `listing_mode` check (`fixed_price_open_claim`, `fixed_price_bid_select`, `reserve_price_auction`, `redundant_competition`, `decomposed_shard_market`)

### 013 `market_claims_bids_and_reputation`

Files:
- `migrations/013_market_claims_bids_and_reputation.up.sql`
- `migrations/013_market_claims_bids_and_reputation.down.sql`

Creates:
- `market_claims`
- `market_bids`
- `market_profile_reputation_snapshots`

Core constraints:
- `claim_status` lifecycle checks
- `bid_status` lifecycle checks
- uniqueness and active-claim protections for incompatible listing modes

### 014 `escrow_and_payout_core`

Files:
- `migrations/014_escrow_and_payout_core.up.sql`
- `migrations/014_escrow_and_payout_core.down.sql`

Creates:
- `funding_accounts`
- `task_escrows`
- `escrow_transfers`
- `payout_accounts`
- `payout_requests`
- `payout_transfers`

Core constraints:
- escrow state checks
- payout state checks
- amount non-negativity and currency constraints
- explicit link from market listing to escrow state

### 015 `market_disputes_and_read_models`

Files:
- `migrations/015_market_disputes_and_read_models.up.sql`
- `migrations/015_market_disputes_and_read_models.down.sql`

Creates:
- `market_disputes`
- `market_dispute_decisions`
- `rm_market_listing_feed`
- `rm_market_claims`
- `rm_market_reputation`
- `rm_market_disputes`
- `rm_market_escrow_state`
- `rm_market_payout_status`

## Backfill/Post-Deploy Work

- seed initial market fee schedules
- seed work-class policy defaults
- bootstrap requester and executor profile records for opted-in tenants/operators

## Verification Checklist

- market mode remains disabled behind policy until rollout approval
- escrow and payout tables remain isolated from internal v1 ledger truth
- listing publication requires funded reserve checks
- payout retry and failure handling do not corrupt escrow balances
