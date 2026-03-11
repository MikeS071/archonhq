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

## Default task workspace rules
- ephemeral
- isolated from long-lived Hermes personal memory
- no unrestricted network by default
- no implicit filesystem reuse by default
