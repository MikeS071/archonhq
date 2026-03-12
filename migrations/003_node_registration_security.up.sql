CREATE TABLE IF NOT EXISTS node_registration_challenges (
  challenge_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  operator_id TEXT NOT NULL REFERENCES operators(operator_id),
  requested_runtime TEXT NOT NULL,
  challenge_nonce TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_node_reg_challenges_tenant_status
  ON node_registration_challenges (tenant_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS node_credentials (
  credential_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  node_id TEXT NOT NULL REFERENCES nodes(node_id),
  token_hash TEXT NOT NULL,
  status TEXT NOT NULL,
  issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_node_credentials_unique_hash
  ON node_credentials (tenant_id, node_id, token_hash);

CREATE INDEX IF NOT EXISTS idx_node_credentials_active
  ON node_credentials (tenant_id, node_id, status);
