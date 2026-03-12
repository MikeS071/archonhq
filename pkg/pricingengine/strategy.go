package pricingengine

import (
	"fmt"
	"strings"

	"github.com/MikeS071/archonhq/pkg/scoring"
)

type QuoteRequest struct {
	TaskID         string  `json:"task_id"`
	BaseRate       float64 `json:"base_rate"`
	BidAdjustment  float64 `json:"bid_adjustment"`
	PredictedRawJW float64 `json:"predicted_raw_jw"`
	QualityFactor  float64 `json:"quality_factor"`
	ReliabilityRF  float64 `json:"reliability_rf"`
	ReserveRatio   float64 `json:"reserve_ratio"`
}

type QuoteResult struct {
	StrategyName     string         `json:"strategy_name"`
	RateValue        float64        `json:"rate_value"`
	RewardMultiplier float64        `json:"reward_multiplier"`
	CreditedJW       float64        `json:"credited_jw"`
	EstimatedGross   float64        `json:"estimated_gross"`
	EstimatedReserve float64        `json:"estimated_reserve"`
	EstimatedNet     float64        `json:"estimated_net"`
	Metadata         map[string]any `json:"metadata"`
}

type RateCard struct {
	TaskFamily  string  `json:"task_family"`
	BaseRate    float64 `json:"base_rate"`
	ReserveHint float64 `json:"reserve_hint"`
}

type Strategy interface {
	Name() string
	Quote(QuoteRequest) (QuoteResult, error)
}

type FixedPlusBid struct{}

func (s FixedPlusBid) Name() string { return "fixed_plus_bid" }
func (s FixedPlusBid) Quote(req QuoteRequest) (QuoteResult, error) {
	if strings.TrimSpace(req.TaskID) == "" {
		return QuoteResult{}, fmt.Errorf("task_id is required")
	}
	rate := req.BaseRate + req.BidAdjustment
	if rate <= 0 {
		return QuoteResult{}, fmt.Errorf("resolved rate must be positive")
	}

	quality := req.QualityFactor
	if quality <= 0 {
		quality = 1
	}
	raw := req.PredictedRawJW
	if raw < 0 {
		raw = 0
	}

	reward := scoring.RewardMultiplierFromRF(req.ReliabilityRF)
	credited := raw * quality * reward
	reserveRatio := req.ReserveRatio
	if reserveRatio < 0 {
		reserveRatio = 0
	}
	if reserveRatio > 1 {
		reserveRatio = 1
	}

	gross := credited * rate
	reserve := gross * reserveRatio
	net := gross - reserve

	return QuoteResult{
		StrategyName:     s.Name(),
		RateValue:        rate,
		RewardMultiplier: reward,
		CreditedJW:       credited,
		EstimatedGross:   gross,
		EstimatedReserve: reserve,
		EstimatedNet:     net,
		Metadata: map[string]any{
			"base_rate":      req.BaseRate,
			"bid_adjustment": req.BidAdjustment,
			"reserve_ratio":  reserveRatio,
		},
	}, nil
}

func ResolveStrategy(name string) Strategy {
	if strings.EqualFold(strings.TrimSpace(name), "fixed_plus_bid") || strings.TrimSpace(name) == "" {
		return FixedPlusBid{}
	}
	return FixedPlusBid{}
}

func DefaultRateCards() []RateCard {
	return []RateCard{
		{TaskFamily: "research.extract", BaseRate: 1.8, ReserveHint: 0.15},
		{TaskFamily: "doc.section.write", BaseRate: 1.6, ReserveHint: 0.10},
		{TaskFamily: "code.patch", BaseRate: 2.4, ReserveHint: 0.20},
		{TaskFamily: "verify.result", BaseRate: 1.5, ReserveHint: 0.10},
		{TaskFamily: "reduce.merge", BaseRate: 1.9, ReserveHint: 0.15},
		{TaskFamily: "autosearch.self_improve", BaseRate: 3.0, ReserveHint: 0.25},
	}
}
