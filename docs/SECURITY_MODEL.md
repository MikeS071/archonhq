# SECURITY_MODEL.md

## Goals
- protect tenant isolation
- protect secrets and provider keys
- protect sensitive task data
- constrain worker execution
- ensure auditability

## Required controls
- TLS everywhere
- encrypted secrets
- tenant-scoped credentials
- signed node registration challenge
- signed result submissions
- isolated ephemeral task workspaces
- network policy per lease
- tool grants per lease
- object storage namespace isolation
- audit logs for privileged actions
- simulation run budget/time limits
- replay approval and redaction for sensitive traces
- simulation namespace isolation from production truth

## Default task workspace rules
- ephemeral
- isolated from long-lived Hermes personal memory
- no unrestricted network by default
- no implicit filesystem reuse by default

## Simulation-specific rules
- simulation runs may use only synthetic, anonymized, or explicitly approved replay data
- simulation runs must not use production provider credentials by default
- simulation artifacts, metrics, and events must use dedicated namespaces
- runtime-backed simulation must run with explicit allowlisted egress only
- baseline promotion and production-trace replay must be audit logged
