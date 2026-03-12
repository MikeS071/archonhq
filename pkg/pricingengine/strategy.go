package pricingengine

import "fmt"

type QuoteRequest struct{}
type QuoteResult struct{}

type Strategy interface {
	Name() string
	Quote(QuoteRequest) (QuoteResult, error)
}

type FixedPlusBid struct{}

func (s FixedPlusBid) Name() string { return "fixed_plus_bid" }
func (s FixedPlusBid) Quote(QuoteRequest) (QuoteResult, error) {
	return QuoteResult{}, fmt.Errorf("not implemented")
}
