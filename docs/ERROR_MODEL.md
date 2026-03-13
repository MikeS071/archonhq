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
- acceptance_contract_invalid
- acceptance_contract_required
- approval_required
- approval_denied
- lease_not_found
- lease_expired
- policy_violation
- artifact_not_found
- verification_failed
- verification_not_found
- critic_registry_not_found
- validation_run_not_found
- validation_stage_failed
- validation_tier_not_allowed
- acceptance_evidence_missing
- reduction_failed
- reduction_not_found
- merge_strategy_unsupported
- task_decompose_failed
- task_market_failed
- market_profile_not_found
- market_listing_not_found
- market_work_class_forbidden
- market_funding_required
- market_claim_conflict
- escrow_not_found
- escrow_insufficient_funds
- payout_not_found
- payout_failed
- dispute_not_found
- dispute_resolution_failed
- simulation_not_found
- simulation_run_failed
- simulation_replay_requires_approval
- simulation_budget_exceeded
- simulation_isolation_violation
- integration_sync_failed
- integration_status_failed
- insufficient_budget
- invalid_signature
- node_registration_failed
- rate_resolution_failed
- settlement_failed
