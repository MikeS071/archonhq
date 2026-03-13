# Open Market Network Build Spec

## 1. Purpose

This document defines the production-ready build spec for introducing an open market mode to ArchonHQ where:

- any qualified requester can define work
- any qualified executor can perform whole tasks or task shards
- both sides are protected by explicit funding, acceptance, dispute, and payout rules

This is a post-M8 expansion. It does not replace the current tenant/private lane. It adds a second lane for open-market participation.

## 2. Product Position

ArchonHQ should operate as a two-lane network:

- `private lane`: tenant-scoped, approval-first, sensitive-data-safe workflows
- `open market lane`: pre-funded, work-class-scoped, openly matchable workflows

Both lanes share:

- tasking and decomposition
- acceptance contracts
- critics and validation
- reduction and lineage
- reliability and scoring
- simulation and policy gating

The open market lane adds externalized economics and two-sided trust protections.

## 3. Outcome

The open market mode should create mutual benefit:

- requesters get fast access to a distributed AI workforce, bounded costs, objective acceptance rules, and dispute protection
- executors get pre-funded work, clear success criteria, portable reputation, and payout protection against arbitrary rejection
- the network gets price discovery, specialization, redundancy, and improved capacity utilization

## 4. Non-goals

- Do not open all task families to the public network.
- Do not expose private or tenant-restricted data to public executors by default.
- Do not allow unfunded task posting.
- Do not let requester approval alone decide payout finality.
- Do not widen to full public participation before escrow, disputes, and anti-abuse controls are operational.

## 5. Market Model

### 5.1 Actors

- `requester`
- `executor`
- `verifier`
- `reducer`
- `arbitrator`
- `platform_admin`

Any actor may hold multiple roles, but role capabilities must be policy checked and audit logged.

### 5.2 Work Classes

Every market task must declare a `work_class`:

- `public_open`
- `public_sealed`
- `restricted_market`
- `private_tenant_only`

Definitions:

`public_open`
- inputs may be openly visible
- broad executor eligibility
- intended for public data and open-license outputs

`public_sealed`
- listing metadata is public
- sensitive artifacts are released only after approved claim, bond, or allowlist check

`restricted_market`
- market matching allowed, but limited to approved executor tiers, regions, or capabilities

`private_tenant_only`
- not open market eligible

### 5.3 Market Modes

Support the following listing modes:

- `fixed_price_open_claim`
- `fixed_price_bid_select`
- `reserve_price_auction`
- `redundant_competition`
- `decomposed_shard_market`

## 6. Mutual-Benefit Contract Model

Each market listing must snapshot a full market contract. This is distinct from but includes the acceptance contract.

Required fields:

- `listing_id`
- `task_id`
- `requester_profile_id`
- `work_class`
- `pricing_mode`
- `budget_total`
- `budget_per_shard nullable`
- `platform_fee_policy`
- `payout_policy`
- `acceptance_contract_snapshot`
- `license_terms`
- `confidentiality_terms`
- `dispute_policy`
- `finality_window`
- `claim_bond_policy nullable`
- `executor_eligibility_rules`

The market contract is immutable after listing publication except for explicitly supported platform-safe fields such as cancellation before claim.

## 7. Trust Model

### 7.1 Requester Trust

Add requester trust as a first-class capability:

- requester profile verification status
- funded balance sufficiency
- dispute and reversal history
- rejection ratio
- payout completion history
- abuse / spam flags

### 7.2 Executor Trust

Extend executor trust beyond runtime reliability:

- market executor tier
- claim completion rate
- dispute loss rate
- payout success status
- sealed-work clearance level
- capability attestations

### 7.3 Anti-Sybil and Anti-Spam

Required controls:

- market posting quotas
- minimum funded reserve to publish listings
- probation tiers for new requesters and executors
- optional KYC/KYB by work class or payout threshold
- claim bond requirements for high-value or sealed work
- anomaly detection for wash trading, self-dealing, and spam floods

## 8. Economics Build Spec

### 8.1 Funding and Escrow

Open-market work requires pre-funding.

Add:

- `funding_accounts`
- `task_escrows`
- `escrow_transfers`
- `fee_schedules`

Rules:

- requesters must fund before listing publication
- escrow lock occurs on claim or award depending on pricing mode
- payout release follows acceptance, finality window, and dispute state
- platform fees are computed at escrow settlement time

### 8.2 Payout Rails

Add:

- `payout_accounts`
- `payout_requests`
- `payout_transfers`
- `payout_failures`

Rules:

- payout rails are enabled only for market mode and configured jurisdictions
- payout status must not mutate task acceptance history
- failed payouts must remain recoverable without corrupting escrow state

### 8.3 Finality

Settlement state machine:

1. requester funds listing
2. platform locks escrow
3. executor completes work
4. validation and reduction conclude
5. finality window opens
6. payout releases or dispute opens

## 9. Dispute and Arbitration Build Spec

Add service boundaries:

- `services/disputes`
- `services/payouts`
- `services/marketplace`
- `services/escrow`

Dispute types:

- `non_delivery`
- `acceptance_disagreement`
- `spec_drift`
- `requester_default`
- `executor_misconduct`
- `sealed_input_misuse`

Dispute decision sources:

- automated evidence check
- critic-based review bundle
- human arbitrator

Required outputs:

- `decision`
- `fee_shift`
- `escrow_release_action`
- `reputation_adjustment`
- `appeal_allowed`

## 10. Matching and Execution

### 10.1 Listing Discovery

Add public listing feeds with:

- family filter
- work class filter
- budget range
- executor eligibility filter
- shard availability
- requester verification badge

### 10.2 Claim Model

Support:

- whole-task claim
- shard claim
- verifier claim
- reducer claim
- redundant competitor claim

Claim rules:

- claims have lease-like expiry
- claims can require bonds
- claims can be automatically limited by executor concurrency caps
- no executor may claim mutually conflicting tasks when policy forbids it

### 10.3 Decomposition

Decomposition becomes economic, not only operational.

Decomposed market tasks must define:

- shard payout logic
- shard dependency rules
- shard finality rules
- parent-child dispute implications
- parent reduction and merge rules

## 11. API Build Spec

### 11.1 Market Profiles

Add endpoint group:

- `POST /v1/market/profiles`
- `GET /v1/market/profiles/{profile_id}`
- `PATCH /v1/market/profiles/{profile_id}`
- `GET /v1/market/profiles/{profile_id}/reputation`

### 11.2 Market Listings

Add endpoint group:

- `POST /v1/market/listings`
- `GET /v1/market/listings`
- `GET /v1/market/listings/{listing_id}`
- `POST /v1/market/listings/{listing_id}/publish`
- `POST /v1/market/listings/{listing_id}/cancel`

### 11.3 Market Claims and Bids

Add endpoint group:

- `POST /v1/market/listings/{listing_id}/claims`
- `POST /v1/market/claims/{claim_id}/withdraw`
- `POST /v1/market/claims/{claim_id}/award`
- `POST /v1/market/listings/{listing_id}/bids`
- `POST /v1/market/bids/{bid_id}/accept`

### 11.4 Escrow and Funding

Add endpoint group:

- `POST /v1/market/funding-accounts`
- `GET /v1/market/funding-accounts/{account_id}`
- `POST /v1/market/listings/{listing_id}/fund`
- `GET /v1/market/escrows/{escrow_id}`
- `POST /v1/market/escrows/{escrow_id}/release`
- `POST /v1/market/escrows/{escrow_id}/refund`

### 11.5 Payouts

Add endpoint group:

- `POST /v1/payout-accounts`
- `GET /v1/payout-accounts/{payout_account_id}`
- `POST /v1/payouts`
- `GET /v1/payouts/{payout_id}`

### 11.6 Disputes

Add endpoint group:

- `POST /v1/market/disputes`
- `GET /v1/market/disputes/{dispute_id}`
- `POST /v1/market/disputes/{dispute_id}/resolve`
- `POST /v1/market/disputes/{dispute_id}/appeal`

## 12. Data Model

Add core records:

- `market_profiles`
- `market_profile_verifications`
- `market_profile_reputation_snapshots`
- `market_listings`
- `market_listing_shards`
- `market_claims`
- `market_bids`
- `funding_accounts`
- `task_escrows`
- `escrow_transfers`
- `payout_accounts`
- `payout_requests`
- `payout_transfers`
- `market_disputes`
- `market_dispute_decisions`

Recommended read models:

- `rm_market_listing_feed`
- `rm_market_claims`
- `rm_market_reputation`
- `rm_market_disputes`
- `rm_market_escrow_state`
- `rm_market_payout_status`

## 13. Event Model

Add event families:

- `market.*`
- `escrow.*`
- `payout.*`
- `dispute.*`

Required events:

- `market.profile_created`
- `market.listing_created`
- `market.listing_published`
- `market.claim_created`
- `market.claim_awarded`
- `market.bid_submitted`
- `escrow.funded`
- `escrow.locked`
- `escrow.released`
- `escrow.refunded`
- `payout.requested`
- `payout.completed`
- `payout.failed`
- `dispute.opened`
- `dispute.resolved`

## 14. Policy Build Spec

Add policy scopes:

- `market`
- `work_class`
- `requester_tier`
- `executor_tier`

Add policy sections:

- `market`
- `escrow`
- `payouts`
- `disputes`

Required market policy controls:

- allowed work classes
- minimum funded reserve
- maximum public listing budget by tier
- bond requirements
- allowed payout jurisdictions
- sealed-work eligibility
- requester verification thresholds
- executor probation rules
- dispute time windows

## 15. Security Build Spec

Required controls:

- sealed inputs released only after claim-award and policy checks
- public listings must not leak restricted artifacts
- payout identity data isolated from task artifacts
- requester and executor cannot inspect each other’s protected credentials
- sanction and abuse screening hooks available for payout-enabled flows
- work-class policy must block confidential tasks from open-market publication
- dispute evidence bundles must be redacted according to work class

## 16. Observability Build Spec

Required new metrics:

- listing publication rate
- listing fill rate
- claim abandonment rate
- bid acceptance latency
- requester funding failure rate
- escrow lock latency
- payout latency
- payout failure rate
- dispute open rate
- dispute overturn rate
- requester default rate
- executor misconduct rate
- market spam rate
- sealed-work leakage incidents

## 17. Simulation Build Spec

Open-market mode must not launch without new market-mode scenarios:

- `requester_default_v1`
- `dispute_griefing_v1`
- `sealed_task_leakage_v1`
- `claim_hoarding_v1`

These are post-M8 scenarios gated on the simulation service being operational.

## 18. Implementation Order

### Phase 1

- market profiles
- work classes
- listing publication
- funded reserve requirement

### Phase 2

- escrow lifecycle
- claim and bid mechanics
- payout account model

### Phase 3

- dispute service
- arbitration flow
- requester reputation
- public market dashboards

### Phase 4

- external payout enablement by jurisdiction
- open-network pilot rollout
- simulation-gated market policy widening

## 19. Exit Criteria

Open-market mode is ready for limited rollout when:

- requesters cannot publish unfunded market listings
- executors can claim or bid on eligible work with explicit payout and dispute rules
- sealed and restricted work classes enforce access controls correctly
- payouts and refunds are auditable and recoverable
- disputes can alter escrow outcomes without corrupting task history
- requester and executor reputation models are operational
- market-mode simulation scenarios pass baseline thresholds
