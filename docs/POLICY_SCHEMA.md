# POLICY_SCHEMA.md

## Policy scopes
- tenant
- workspace
- family
- task
- node
- provider
- market
- work_class
- requester_tier
- executor_tier
- sensitivity tier
- simulation_scenario

## Policy sections
- approval
- execution
- validation
- market
- escrow
- payouts
- disputes
- pricing
- retention
- security
- reliability backoff
- simulation

## Example policy JSON
```json
{
  "scope": "workspace",
  "scope_id": "ws_01",
  "approval": {
    "mode": "always_required"
  },
  "validation": {
    "default_tier": "standard",
    "require_acceptance_contract": true,
    "allow_inline_contracts": true,
    "allow_contract_overrides": false,
    "critic_diversity_mode": "enforced_for_high_assurance",
    "require_cross_provider_critics": false,
    "max_stage_retries": 3
  },
  "market": {
    "enabled": false,
    "allowed_work_classes": ["public_open"],
    "minimum_funded_reserve_usd": 50,
    "allow_public_listing": false,
    "require_requester_verification_for_high_value": true
  },
  "escrow": {
    "required_for_market_mode": true,
    "finality_window_hours": 24,
    "claim_bond_mode": "risk_based"
  },
  "payouts": {
    "enabled": false,
    "allowed_jurisdictions": [],
    "manual_review_threshold_usd": 1000
  },
  "disputes": {
    "enabled": true,
    "default_open_window_hours": 24,
    "appeal_window_hours": 48
  },
  "execution": {
    "allowed_backends": ["docker", "ssh"],
    "allowed_toolsets": ["file", "terminal"],
    "network_policy": "restricted"
  },
  "pricing": {
    "mode": "fixed_plus_bid"
  },
  "retention": {
    "workspace_default": "ephemeral"
  },
  "security": {
    "allow_raw_tool_output_to_models": false,
    "require_evidence_redaction_for_validation": true
  },
  "simulation": {
    "enabled": true,
    "allowed_run_modes": ["deterministic_stub", "sampled_synthetic"],
    "requires_approval_for_replay": true
  }
}
```
