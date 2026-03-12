package settlement

import "testing"

func TestDefaultEngine(t *testing.T) {
	e := DefaultEngine{}
	got, err := e.Settle(ScoreInput{
		Metering: MeteringInput{
			CPUSec:   100,
			GPUClass: "cpu-only",
		},
		TaskDifficulty: "standard",
		Quality: QualityInput{
			Validity:         1,
			VerifierScore:    1,
			AcceptanceSignal: 1,
			Novelty:          1,
			LatencyScore:     1,
		},
		RFLast100:    0.95,
		RFLast30d:    0.95,
		RFLifetime:   0.95,
		Rate:         2.0,
		ReserveRatio: 0.1,
	})
	if err != nil {
		t.Fatalf("settle error: %v", err)
	}

	if got.RawJW < 0.2-0.000001 || got.RawJW > 0.2+0.000001 {
		t.Fatalf("unexpected raw_jw: %f", got.RawJW)
	}
	if got.QualityFactor < 1-0.000001 || got.QualityFactor > 1+0.000001 {
		t.Fatalf("unexpected quality factor: %f", got.QualityFactor)
	}
	if got.RFFinal < 0.95-0.000001 || got.RFFinal > 0.95+0.000001 {
		t.Fatalf("unexpected rf_final: %f", got.RFFinal)
	}
	if got.RewardMultiplier < 1-0.000001 || got.RewardMultiplier > 1+0.000001 {
		t.Fatalf("unexpected reward multiplier: %f", got.RewardMultiplier)
	}
	if got.CreditedJW < 0.2-0.000001 || got.CreditedJW > 0.2+0.000001 {
		t.Fatalf("unexpected credited_jw: %f", got.CreditedJW)
	}
	if got.GrossAmount < 0.4-0.000001 || got.GrossAmount > 0.4+0.000001 {
		t.Fatalf("unexpected gross amount: %f", got.GrossAmount)
	}
	if got.ReserveAmount < 0.04-0.000001 || got.ReserveAmount > 0.04+0.000001 {
		t.Fatalf("unexpected reserve amount: %f", got.ReserveAmount)
	}
	if got.NetAmount < 0.36-0.000001 || got.NetAmount > 0.36+0.000001 {
		t.Fatalf("unexpected net amount: %f", got.NetAmount)
	}
}

func TestDefaultEngineValidation(t *testing.T) {
	e := DefaultEngine{}
	if _, err := e.Settle(ScoreInput{}); err == nil {
		t.Fatalf("expected validation error")
	}
}
