# POLICY_SCHEMA.md

## Policy scopes
- tenant
- workspace
- family
- task
- node
- provider
- sensitivity tier

## Policy sections
- approval
- execution
- pricing
- retention
- security
- reliability backoff

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
  }
}
```
