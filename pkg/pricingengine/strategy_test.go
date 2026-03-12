package pricingengine

import "testing"

func TestFixedPlusBid(t *testing.T) {
	s := FixedPlusBid{}
	if s.Name() != "fixed_plus_bid" {
		t.Fatalf("unexpected strategy name")
	}
	got, err := s.Quote(QuoteRequest{
		TaskID:         "task_01",
		BaseRate:       2.0,
		BidAdjustment:  0.5,
		PredictedRawJW: 10.0,
		QualityFactor:  0.8,
		ReliabilityRF:  0.9,
		ReserveRatio:   0.15,
	})
	if err != nil {
		t.Fatalf("quote error: %v", err)
	}
	if got.StrategyName != "fixed_plus_bid" {
		t.Fatalf("unexpected strategy in quote result: %s", got.StrategyName)
	}
	if got.RateValue != 2.5 {
		t.Fatalf("unexpected rate value: %f", got.RateValue)
	}
	if got.CreditedJW < 7.36-0.000001 || got.CreditedJW > 7.36+0.000001 {
		t.Fatalf("unexpected credited_jw: %f", got.CreditedJW)
	}
	if got.EstimatedGross < 18.4-0.000001 || got.EstimatedGross > 18.4+0.000001 {
		t.Fatalf("unexpected gross: %f", got.EstimatedGross)
	}
	if got.EstimatedReserve < 2.76-0.000001 || got.EstimatedReserve > 2.76+0.000001 {
		t.Fatalf("unexpected reserve: %f", got.EstimatedReserve)
	}
	if got.EstimatedNet < 15.64-0.000001 || got.EstimatedNet > 15.64+0.000001 {
		t.Fatalf("unexpected net: %f", got.EstimatedNet)
	}
}

func TestFixedPlusBidValidation(t *testing.T) {
	s := FixedPlusBid{}
	if _, err := s.Quote(QuoteRequest{TaskID: "", BaseRate: 1}); err == nil {
		t.Fatalf("expected validation error for missing task_id")
	}
	if _, err := s.Quote(QuoteRequest{TaskID: "task_01", BaseRate: -1}); err == nil {
		t.Fatalf("expected validation error for non-positive rate")
	}
}

func TestResolveStrategyAndRateCards(t *testing.T) {
	if got := ResolveStrategy("").Name(); got != "fixed_plus_bid" {
		t.Fatalf("unexpected default strategy: %s", got)
	}
	if got := ResolveStrategy("fixed_plus_bid").Name(); got != "fixed_plus_bid" {
		t.Fatalf("unexpected explicit strategy: %s", got)
	}
	if got := ResolveStrategy("unknown").Name(); got != "fixed_plus_bid" {
		t.Fatalf("unexpected fallback strategy: %s", got)
	}

	cards := DefaultRateCards()
	if len(cards) == 0 {
		t.Fatalf("expected non-empty default rate cards")
	}
	for _, card := range cards {
		if card.TaskFamily == "" {
			t.Fatalf("task family must be set")
		}
		if card.BaseRate <= 0 {
			t.Fatalf("base rate must be positive")
		}
	}
}
