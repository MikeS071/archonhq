# TEST_PLAN.md

## Unit tests
- JouleWork formula math
- quality scoring
- reliability scoring
- reward multipliers
- pricing strategy selection
- settlement posting logic
- simulation metric and finding derivation

## Integration tests
- tenant/workspace lifecycle
- task -> approval -> lease -> result -> verification -> reduction -> settlement
- node registration
- artifact upload/register flow
- policy enforcement
- reserve hold lifecycle
- simulation scenario -> run -> findings -> baseline comparison
- simulation isolation from production tables and read models

## Contract tests
- Hermes adapter
- Paperclip connector
- simulation scenario registry and run APIs

## Security tests
- tenant isolation
- forbidden access checks
- invalid signature rejection
- replay approval enforcement for sensitive traces
- simulation namespace isolation
