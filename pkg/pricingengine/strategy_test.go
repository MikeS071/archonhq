package pricingengine

import "testing"

func TestFixedPlusBid(t *testing.T) {
	s := FixedPlusBid{}
	if s.Name() != "fixed_plus_bid" {
		t.Fatalf("unexpected strategy name")
	}
	if _, err := s.Quote(QuoteRequest{}); err == nil {
		t.Fatalf("expected not implemented error")
	}
}
