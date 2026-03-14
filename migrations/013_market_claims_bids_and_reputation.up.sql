CREATE TABLE IF NOT EXISTS market_claims (
  claim_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  listing_id TEXT NOT NULL REFERENCES market_listings(listing_id),
  executor_profile_id TEXT NOT NULL,
  claim_type TEXT NOT NULL CHECK (claim_type IN ('whole_task', 'shard', 'verifier', 'reducer', 'redundant_competitor')),
  bond_amount NUMERIC(18,8) NOT NULL DEFAULT 0 CHECK (bond_amount >= 0),
  status TEXT NOT NULL CHECK (status IN ('submitted', 'withdrawn', 'awarded', 'rejected')),
  policy_checks_passed BOOLEAN NOT NULL DEFAULT FALSE,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, executor_profile_id) REFERENCES market_profiles(tenant_id, profile_id)
);

CREATE TABLE IF NOT EXISTS market_bids (
  bid_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  listing_id TEXT NOT NULL REFERENCES market_listings(listing_id),
  executor_profile_id TEXT NOT NULL,
  amount NUMERIC(18,8) NOT NULL CHECK (amount > 0),
  currency TEXT NOT NULL DEFAULT 'JWUSD',
  status TEXT NOT NULL CHECK (status IN ('submitted', 'accepted', 'rejected')),
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, executor_profile_id) REFERENCES market_profiles(tenant_id, profile_id)
);

CREATE TABLE IF NOT EXISTS market_profile_reputation_snapshots (
  snapshot_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  profile_id TEXT NOT NULL,
  claim_completion_rate DOUBLE PRECISION,
  dispute_loss_rate DOUBLE PRECISION,
  payout_success_rate DOUBLE PRECISION,
  rejection_ratio DOUBLE PRECISION,
  score DOUBLE PRECISION,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  as_of TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, profile_id) REFERENCES market_profiles(tenant_id, profile_id)
);

CREATE INDEX IF NOT EXISTS idx_market_claims_tenant_listing_status ON market_claims(tenant_id, listing_id, status);
CREATE INDEX IF NOT EXISTS idx_market_bids_tenant_listing_status ON market_bids(tenant_id, listing_id, status);
CREATE INDEX IF NOT EXISTS idx_market_reputation_tenant_profile ON market_profile_reputation_snapshots(tenant_id, profile_id, as_of DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_market_claim_executor_active_per_listing
  ON market_claims(tenant_id, listing_id, executor_profile_id)
  WHERE status IN ('submitted', 'awarded');
