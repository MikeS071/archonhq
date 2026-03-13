# POLICY_SCHEMA.md

## Policy scopes
- tenant
- workspace
- family
- task
- node
- provider
- sensitivity tier
- simulation_scenario

## Policy sections
- approval
- execution
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
  "simulation": {
    "enabled": true,
    "allowed_run_modes": ["deterministic_stub", "sampled_synthetic"],
    "requires_approval_for_replay": true
  }
}
```
