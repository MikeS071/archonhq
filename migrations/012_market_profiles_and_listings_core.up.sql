CREATE TABLE IF NOT EXISTS market_profiles (
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  profile_id TEXT NOT NULL,
  profile_type TEXT NOT NULL CHECK (profile_type IN ('requester', 'executor', 'hybrid')),
  display_name TEXT NOT NULL,
  verification_status TEXT NOT NULL DEFAULT 'pending' CHECK (verification_status IN ('pending', 'verified', 'rejected')),
  executor_tier TEXT,
  status TEXT NOT NULL DEFAULT 'active',
  capability_tags_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  region_allowlist_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  work_class_allowlist_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, profile_id)
);

CREATE TABLE IF NOT EXISTS market_profile_verifications (
  market_profile_verification_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  profile_id TEXT NOT NULL,
  verification_source TEXT NOT NULL,
  verification_status TEXT NOT NULL CHECK (verification_status IN ('pending', 'verified', 'rejected')),
  evidence_ref TEXT,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, profile_id) REFERENCES market_profiles(tenant_id, profile_id)
);

CREATE TABLE IF NOT EXISTS market_listings (
  listing_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  task_id TEXT NOT NULL REFERENCES tasks(task_id),
  requester_profile_id TEXT NOT NULL,
  work_class TEXT NOT NULL CHECK (work_class IN ('public_open', 'public_sealed', 'restricted_market', 'private_tenant_only')),
  listing_mode TEXT NOT NULL CHECK (listing_mode IN ('fixed_price_open_claim', 'fixed_price_bid_select', 'reserve_price_auction', 'redundant_competition', 'decomposed_shard_market')),
  budget_total NUMERIC(18,8) NOT NULL CHECK (budget_total > 0),
  budget_per_shard NUMERIC(18,8),
  currency TEXT NOT NULL DEFAULT 'JWUSD',
  funding_account_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'cancelled', 'completed')),
  contract_snapshot_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  publish_reason TEXT,
  cancel_reason TEXT,
  published_at TIMESTAMPTZ,
  cancelled_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, requester_profile_id) REFERENCES market_profiles(tenant_id, profile_id)
);

CREATE TABLE IF NOT EXISTS market_listing_shards (
  market_listing_shard_id TEXT PRIMARY KEY,
  listing_id TEXT NOT NULL REFERENCES market_listings(listing_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  shard_key TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'claimed', 'completed', 'cancelled')),
  budget NUMERIC(18,8) NOT NULL CHECK (budget >= 0),
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (listing_id, shard_key)
);

CREATE INDEX IF NOT EXISTS idx_market_profiles_tenant_type ON market_profiles(tenant_id, profile_type);
CREATE INDEX IF NOT EXISTS idx_market_listings_tenant_status ON market_listings(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_market_listings_work_class ON market_listings(work_class, listing_mode);
