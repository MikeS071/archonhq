CREATE TABLE IF NOT EXISTS market_disputes (
  dispute_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  listing_id TEXT NOT NULL REFERENCES market_listings(listing_id),
  escrow_id TEXT REFERENCES task_escrows(escrow_id),
  claim_id TEXT REFERENCES market_claims(claim_id),
  dispute_type TEXT NOT NULL CHECK (dispute_type IN ('non_delivery', 'acceptance_disagreement', 'spec_drift', 'requester_default', 'executor_misconduct', 'sealed_input_misuse')),
  reason TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('open', 'resolved', 'appealed')),
  opened_by TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS market_dispute_decisions (
  dispute_decision_id TEXT PRIMARY KEY,
  dispute_id TEXT NOT NULL REFERENCES market_disputes(dispute_id),
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  decision TEXT NOT NULL,
  fee_shift DOUBLE PRECISION,
  escrow_release_action TEXT NOT NULL CHECK (escrow_release_action IN ('none', 'release', 'refund')),
  reputation_adjustment_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  appeal_allowed BOOLEAN NOT NULL DEFAULT FALSE,
  resolved_by TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rm_market_listing_feed (
  listing_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  work_class TEXT NOT NULL,
  listing_mode TEXT NOT NULL,
  status TEXT NOT NULL,
  budget_total NUMERIC(18,8) NOT NULL,
  currency TEXT NOT NULL,
  requester_profile_id TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rm_market_claims (
  claim_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  listing_id TEXT NOT NULL,
  executor_profile_id TEXT NOT NULL,
  status TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rm_market_reputation (
  tenant_id TEXT NOT NULL,
  profile_id TEXT NOT NULL,
  score DOUBLE PRECISION,
  claim_completion_rate DOUBLE PRECISION,
  dispute_loss_rate DOUBLE PRECISION,
  payout_success_rate DOUBLE PRECISION,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, profile_id)
);

CREATE TABLE IF NOT EXISTS rm_market_disputes (
  dispute_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  listing_id TEXT NOT NULL,
  dispute_type TEXT NOT NULL,
  status TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rm_market_escrow_state (
  escrow_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  listing_id TEXT NOT NULL,
  status TEXT NOT NULL,
  total_locked NUMERIC(18,8) NOT NULL DEFAULT 0,
  released_amount NUMERIC(18,8) NOT NULL DEFAULT 0,
  refunded_amount NUMERIC(18,8) NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rm_market_payout_status (
  payout_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  payout_account_id TEXT NOT NULL,
  status TEXT NOT NULL,
  amount NUMERIC(18,8) NOT NULL DEFAULT 0,
  currency TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_market_disputes_tenant_status ON market_disputes(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_rm_market_listing_feed_tenant_status ON rm_market_listing_feed(tenant_id, status);
