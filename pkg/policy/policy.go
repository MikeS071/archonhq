package policy

type Bundle struct {
	Scope       string         `json:"scope"`
	ScopeID     string         `json:"scope_id"`
	Approval    ApprovalPolicy `json:"approval"`
	Execution   ExecutionRule  `json:"execution"`
	Pricing     PricingRule    `json:"pricing"`
	Retention   RetentionRule  `json:"retention"`
	Security    SecurityRule   `json:"security,omitempty"`
	Reliability BackoffRule    `json:"reliability_backoff,omitempty"`
}

type ApprovalPolicy struct {
	Mode string `json:"mode"`
}

type ExecutionRule struct {
	AllowedBackends []string `json:"allowed_backends"`
	AllowedToolsets []string `json:"allowed_toolsets"`
	NetworkPolicy   string   `json:"network_policy"`
}

type PricingRule struct {
	Mode string `json:"mode"`
}

type RetentionRule struct {
	WorkspaceDefault string `json:"workspace_default"`
}

type SecurityRule struct {
	RequireSignature bool `json:"require_signature"`
}

type BackoffRule struct {
	Enabled bool `json:"enabled"`
}
