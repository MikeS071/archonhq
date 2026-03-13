CREATE TABLE IF NOT EXISTS acceptance_contract_templates (
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  template_id TEXT NOT NULL,
  task_family TEXT NOT NULL,
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
  published_version INT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, template_id)
);

CREATE TABLE IF NOT EXISTS acceptance_contract_template_versions (
  tenant_id TEXT NOT NULL,
  template_id TEXT NOT NULL,
  version INT NOT NULL,
  contract_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, template_id, version),
  FOREIGN KEY (tenant_id, template_id) REFERENCES acceptance_contract_templates(tenant_id, template_id)
);

CREATE TABLE IF NOT EXISTS task_acceptance_contracts (
  contract_id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(tenant_id),
  task_id TEXT NOT NULL REFERENCES tasks(task_id),
  contract_source TEXT NOT NULL CHECK (contract_source IN ('inline', 'template_ref', 'template_ref_with_overrides')),
  template_id TEXT,
  template_version INT,
  validation_tier TEXT NOT NULL CHECK (validation_tier IN ('fast', 'standard', 'high_assurance')),
  contract_snapshot_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, task_id),
  FOREIGN KEY (tenant_id, template_id, template_version)
    REFERENCES acceptance_contract_template_versions(tenant_id, template_id, version)
);
