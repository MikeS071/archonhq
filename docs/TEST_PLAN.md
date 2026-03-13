# TEST_PLAN.md

## Unit tests
- JouleWork formula math
- quality scoring
- reliability scoring
- reward multipliers
- pricing strategy selection
- settlement posting logic
- acceptance contract validation and snapshotting
- validation tier policy enforcement
- critic bundle selection and diversity checks
- escrow and payout state-machine logic
- dispute fee-shift and release-decision logic
- simulation metric and finding derivation

## Integration tests
- tenant/workspace lifecycle
- task -> approval -> lease -> result -> verification -> reduction -> settlement
- task -> acceptance contract attach -> validation run -> stage retry/escalation
- listing -> funding -> claim -> validation -> escrow release -> payout
- listing -> dispute -> arbitration -> refund or payout
- node registration
- artifact upload/register flow
- policy enforcement
- reserve hold lifecycle
- simulation scenario -> run -> findings -> baseline comparison
- simulation isolation from production tables and read models

## Contract tests
- Hermes adapter
- Paperclip connector
- acceptance contract template APIs
- critic registry APIs
- validation run APIs
- market listing, escrow, payout, and dispute APIs
- simulation scenario registry and run APIs

## Security tests
- tenant isolation
- forbidden access checks
- invalid signature rejection
- acceptance contract override authorization
- raw tool output exposure policy enforcement
- public listing blocked for private work class
- sealed-work access checks before award
- replay approval enforcement for sensitive traces
- simulation namespace isolation
