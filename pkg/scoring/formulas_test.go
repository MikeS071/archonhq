package scoring

import "testing"

func TestComputeRawJouleWork(t *testing.T) {
	got := ComputeRawJouleWork(Metering{
		CPUSec:            100,
		GPUSec:            10,
		GPUClass:          "rtx_4090_class",
		TokensIn:          1000,
		TokensOut:         500,
		ExternalToolCalls: 3,
		NetworkMB:         20,
		StorageMB:         50,
	}, "hard")

	want := 0.98
	if got < want-0.000001 || got > want+0.000001 {
		t.Fatalf("raw joulework mismatch got=%f want=%f", got, want)
	}
}

func TestQualityAndReliabilityFormulas(t *testing.T) {
	quality := ComputeQualityScore(QualityInputs{
		Validity:         0.9,
		VerifierScore:    0.8,
		AcceptanceSignal: 0.7,
		Novelty:          0.6,
		LatencyScore:     0.5,
	})
	if quality < 0.78-0.000001 || quality > 0.78+0.000001 {
		t.Fatalf("quality mismatch got=%f want=0.78", quality)
	}

	agent := ComputeAgentRF(AgentReliabilityInputs{
		ValidityRate:         0.9,
		VerificationPassRate: 0.8,
		AcceptanceRate:       0.7,
		RollbackRate:         0.1,
		DisputeLossRate:      0.05,
	})
	if agent < 0.84-0.000001 || agent > 0.84+0.000001 {
		t.Fatalf("agent rf mismatch got=%f want=0.84", agent)
	}

	operator := ComputeOperatorRF(OperatorReliabilityInputs{
		FleetAcceptanceRate:   0.85,
		FleetVerificationRate: 0.8,
		FleetReworkRate:       0.1,
		DisputeRate:           0.05,
		UptimeScore:           0.9,
	})
	if operator < 0.86-0.000001 || operator > 0.86+0.000001 {
		t.Fatalf("operator rf mismatch got=%f want=0.86", operator)
	}

	effective := ComputeEffectiveRF(agent, operator)
	if effective < 0.847-0.000001 || effective > 0.847+0.000001 {
		t.Fatalf("effective rf mismatch got=%f want=0.847", effective)
	}

	final := ComputeFinalRF(0.9, 0.8, 0.7)
	if final < 0.83-0.000001 || final > 0.83+0.000001 {
		t.Fatalf("final rf mismatch got=%f want=0.83", final)
	}
}

func TestRewardMultiplierFromRF(t *testing.T) {
	cases := []struct {
		rf   float64
		want float64
	}{
		{0.96, 1.00},
		{0.90, 0.92},
		{0.80, 0.75},
		{0.70, 0.50},
		{0.55, 0.20},
		{0.54, 0.00},
	}

	for _, tc := range cases {
		if got := RewardMultiplierFromRF(tc.rf); got != tc.want {
			t.Fatalf("reward multiplier mismatch rf=%f got=%f want=%f", tc.rf, got, tc.want)
		}
	}
}

func TestFormulaDefaultsAndClamping(t *testing.T) {
	raw := ComputeRawJouleWork(Metering{
		CPUSec:            1,
		GPUSec:            1,
		GPUClass:          "unknown_gpu",
		TokensIn:          0,
		TokensOut:         0,
		ExternalToolCalls: 0,
		NetworkMB:         0,
		StorageMB:         0,
	}, "unknown")
	if raw < 0.012-0.000001 || raw > 0.012+0.000001 {
		t.Fatalf("unexpected default raw joulework: %f", raw)
	}

	q := ComputeQualityScore(QualityInputs{
		Validity:         2,
		VerifierScore:    2,
		AcceptanceSignal: -1,
		Novelty:          -1,
		LatencyScore:     2,
	})
	if q < 0.70-0.000001 || q > 0.70+0.000001 {
		t.Fatalf("unexpected clamped quality score: %f", q)
	}

	if got := ComputeFinalRF(2, -1, 0.5); got < 0.6-0.000001 || got > 0.6+0.000001 {
		t.Fatalf("unexpected clamped final rf: %f", got)
	}
}
