package settlement

import "fmt"

type ScoreInput struct{}
type Payout struct{}

type Engine interface {
	Settle(input ScoreInput) (Payout, error)
}

type DefaultEngine struct{}

func (e DefaultEngine) Settle(ScoreInput) (Payout, error) {
	return Payout{}, fmt.Errorf("not implemented")
}
