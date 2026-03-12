package settlement

import (
	"fmt"

	"github.com/MikeS071/archonhq/pkg/scoring"
)

type MeteringInput = scoring.Metering
type QualityInput = scoring.QualityInputs

type ScoreInput struct {
	Metering       MeteringInput `json:"metering"`
	TaskDifficulty string        `json:"task_difficulty"`
	Quality        QualityInput  `json:"quality"`
	RFLast100      float64       `json:"rf_last_100"`
	RFLast30d      float64       `json:"rf_last_30d"`
	RFLifetime     float64       `json:"rf_lifetime"`
	Rate           float64       `json:"rate"`
	ReserveRatio   float64       `json:"reserve_ratio"`
}

type Payout struct {
	RawJW            float64 `json:"raw_jw"`
	QualityFactor    float64 `json:"quality_factor"`
	RFFinal          float64 `json:"rf_final"`
	RewardMultiplier float64 `json:"reward_multiplier"`
	CreditedJW       float64 `json:"credited_jw"`
	Rate             float64 `json:"rate"`
	GrossAmount      float64 `json:"gross_amount"`
	ReserveAmount    float64 `json:"reserve_amount"`
	NetAmount        float64 `json:"net_amount"`
}

type Engine interface {
	Settle(input ScoreInput) (Payout, error)
}

type DefaultEngine struct{}

func (e DefaultEngine) Settle(input ScoreInput) (Payout, error) {
	if input.Rate <= 0 {
		return Payout{}, fmt.Errorf("rate must be positive")
	}

	reserveRatio := input.ReserveRatio
	if reserveRatio < 0 {
		reserveRatio = 0
	}
	if reserveRatio > 1 {
		reserveRatio = 1
	}

	rawJW := scoring.ComputeRawJouleWork(input.Metering, input.TaskDifficulty)
	quality := scoring.ComputeQualityScore(input.Quality)
	rfFinal := scoring.ComputeFinalRF(input.RFLast100, input.RFLast30d, input.RFLifetime)
	reward := scoring.RewardMultiplierFromRF(rfFinal)
	credited := rawJW * quality * reward
	gross := credited * input.Rate
	reserve := gross * reserveRatio
	net := gross - reserve

	return Payout{
		RawJW:            rawJW,
		QualityFactor:    quality,
		RFFinal:          rfFinal,
		RewardMultiplier: reward,
		CreditedJW:       credited,
		Rate:             input.Rate,
		GrossAmount:      gross,
		ReserveAmount:    reserve,
		NetAmount:        net,
	}, nil
}
