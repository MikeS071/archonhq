DROP INDEX IF EXISTS uq_market_claim_executor_active_per_listing;
DROP INDEX IF EXISTS idx_market_reputation_tenant_profile;
DROP INDEX IF EXISTS idx_market_bids_tenant_listing_status;
DROP INDEX IF EXISTS idx_market_claims_tenant_listing_status;

DROP TABLE IF EXISTS market_profile_reputation_snapshots;
DROP TABLE IF EXISTS market_bids;
DROP TABLE IF EXISTS market_claims;
