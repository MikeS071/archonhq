# ERROR_MODEL.md

## Error envelope
```json
{
  "error": {
    "code": "approval_required",
    "message": "Task execution requires approval.",
    "details": {
      "approval_request_id": "apr_01"
    },
    "correlation_id": "corr_01"
  }
}
```

## Common error codes
- unauthorized
- forbidden
- tenant_not_found
- workspace_not_found
- task_not_found
- approval_required
- approval_denied
- lease_not_found
- lease_expired
- policy_violation
- artifact_not_found
- verification_failed
- insufficient_budget
- invalid_signature
- node_registration_failed
- rate_resolution_failed
- settlement_failed
