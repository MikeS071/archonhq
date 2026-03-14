CREATE TABLE IF NOT EXISTS funding_accounts (
  account_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  owner_profile_id TEXT NOT NULL,
  currency TEXT NOT NULL DEFAULT 'JWUSD',
  available_balance NUMERIC(18,8) NOT NULL DEFAULT 0 CHECK (available_balance >= 0),
  reserved_balance NUMERIC(18,8) NOT NULL DEFAULT 0 CHECK (reserved_balance >= 0),
  reserve_policy_id TEXT,
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'closed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, owner_profile_id) REFERENCES market_profiles(tenant_id, profile_id)
);

CREATE TABLE IF NOT EXISTS task_escrows (
  escrow_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  listing_id TEXT NOT NULL REFERENCES market_listings(listing_id),
  funding_account_id TEXT NOT NULL REFERENCES funding_accounts(account_id),
  currency TEXT NOT NULL DEFAULT 'JWUSD',
  total_locked NUMERIC(18,8) NOT NULL DEFAULT 0 CHECK (total_locked >= 0),
  released_amount NUMERIC(18,8) NOT NULL DEFAULT 0 CHECK (released_amount >= 0),
  refunded_amount NUMERIC(18,8) NOT NULL DEFAULT 0 CHECK (refunded_amount >= 0),
  status TEXT NOT NULL CHECK (status IN ('pending_lock', 'locked', 'released', 'refunded', 'disputed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS escrow_transfers (
  escrow_transfer_id TEXT PRIMARY KEY,
  escrow_id TEXT NOT NULL REFERENCES task_escrows(escrow_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  transfer_type TEXT NOT NULL CHECK (transfer_type IN ('lock', 'release', 'refund', 'fee')),
  amount NUMERIC(18,8) NOT NULL CHECK (amount >= 0),
  currency TEXT NOT NULL DEFAULT 'JWUSD',
  status TEXT NOT NULL CHECK (status IN ('pending', 'posted', 'failed')),
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payout_accounts (
  payout_account_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  owner_profile_id TEXT NOT NULL,
  provider TEXT NOT NULL,
  provider_account_ref TEXT NOT NULL,
  jurisdiction TEXT,
  status TEXT NOT NULL CHECK (status IN ('pending_verification', 'active', 'suspended', 'closed')),
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, owner_profile_id) REFERENCES market_profiles(tenant_id, profile_id)
);

CREATE TABLE IF NOT EXISTS payout_requests (
  payout_request_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  escrow_id TEXT REFERENCES task_escrows(escrow_id),
  payout_account_id TEXT NOT NULL REFERENCES payout_accounts(payout_account_id),
  amount NUMERIC(18,8) NOT NULL CHECK (amount >= 0),
  currency TEXT NOT NULL DEFAULT 'JWUSD',
  status TEXT NOT NULL CHECK (status IN ('requested', 'approved', 'processing', 'completed', 'failed', 'cancelled')),
  failure_reason TEXT,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payout_transfers (
  payout_transfer_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  payout_request_id TEXT NOT NULL REFERENCES payout_requests(payout_request_id),
  provider_transfer_ref TEXT,
  amount NUMERIC(18,8) NOT NULL CHECK (amount >= 0),
  currency TEXT NOT NULL DEFAULT 'JWUSD',
  status TEXT NOT NULL CHECK (status IN ('pending', 'submitted', 'completed', 'failed')),
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_funding_accounts_tenant_owner ON funding_accounts(tenant_id, owner_profile_id);
CREATE INDEX IF NOT EXISTS idx_task_escrows_listing_status ON task_escrows(listing_id, status);
CREATE INDEX IF NOT EXISTS idx_payout_requests_status ON payout_requests(tenant_id, status);
