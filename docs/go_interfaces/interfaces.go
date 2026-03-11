package interfaces

import "context"

type WorkerAdapter interface {
	Name() string
	Capabilities(ctx context.Context) (NodeCapabilities, error)
	StartLease(ctx context.Context, lease Lease, task TaskSpec) (RunHandle, error)
	PollRun(ctx context.Context, handle RunHandle) (RunStatus, error)
	CollectResult(ctx context.Context, handle RunHandle) (ResultBundle, error)
	CancelRun(ctx context.Context, handle RunHandle) error
}

type MergeStrategy interface {
	Name() string
	ValidateTask(TaskSpec) error
	DetectConflicts(inputs []ResultClaim, basis StateRef) ([]Conflict, error)
	Reduce(inputs []ResultClaim, basis StateRef) (Reduction, error)
}

type Verifier interface {
	Name() string
	Verify(task TaskSpec, result ResultClaim) (VerificationReport, error)
}

type PricingStrategy interface {
	Name() string
	Quote(task TaskSpec, market MarketContext) (PriceQuote, error)
	ResolveRate(task TaskSpec, result ResultClaim, market MarketContext) (RateSnapshot, error)
}

type SettlementEngine interface {
	ComputeRawJW(m Metering, tier TaskTier) float64
	ComputeQuality(q QualityInputs) float64
	ComputeEffectiveRF(r ReliabilitySnapshot) float64
	RewardMultiplier(rf float64) float64
	Score(m Metering, q QualityInputs, r ReliabilitySnapshot, tier TaskTier) (ScoredResult, error)
	Settle(scored ScoredResult, rate float64, reserveFrac float64) (Payout, error)
}
