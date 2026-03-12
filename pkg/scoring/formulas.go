package scoring

import "strings"

type AgentReliabilityInputs struct {
	ValidityRate         float64 `json:"validity_rate"`
	VerificationPassRate float64 `json:"verification_pass_rate"`
	AcceptanceRate       float64 `json:"acceptance_rate"`
	RollbackRate         float64 `json:"rollback_rate"`
	DisputeLossRate      float64 `json:"dispute_loss_rate"`
}

type OperatorReliabilityInputs struct {
	FleetAcceptanceRate   float64 `json:"fleet_acceptance_rate"`
	FleetVerificationRate float64 `json:"fleet_verification_rate"`
	FleetReworkRate       float64 `json:"fleet_rework_rate"`
	DisputeRate           float64 `json:"dispute_rate"`
	UptimeScore           float64 `json:"uptime_score"`
}

func ComputeRawJouleWork(m Metering, taskDifficulty string) float64 {
	cpuComponent := m.CPUSec * 0.002
	gpuComponent := m.GPUSec * gpuWeight(m.GPUClass)
	tokenComponent := (float64(m.TokensIn) + 2*float64(m.TokensOut)) * 0.000002
	toolComponent := float64(m.ExternalToolCalls) * 0.02
	ioComponent := m.NetworkMB*0.0005 + m.StorageMB*0.0002

	raw := cpuComponent + gpuComponent + tokenComponent + toolComponent + ioComponent
	return raw * taskMultiplier(taskDifficulty)
}

func ComputeQualityScore(in QualityInputs) float64 {
	validity := clamp01(in.Validity)
	verifier := clamp01(in.VerifierScore)
	acceptance := clamp01(in.AcceptanceSignal)
	novelty := clamp01(in.Novelty)
	latency := clamp01(in.LatencyScore)

	return 0.35*validity + 0.30*verifier + 0.20*acceptance + 0.10*novelty + 0.05*latency
}

func ComputeAgentRF(in AgentReliabilityInputs) float64 {
	return 0.30*clamp01(in.ValidityRate) +
		0.25*clamp01(in.VerificationPassRate) +
		0.20*clamp01(in.AcceptanceRate) +
		0.15*(1-clamp01(in.RollbackRate)) +
		0.10*(1-clamp01(in.DisputeLossRate))
}

func ComputeOperatorRF(in OperatorReliabilityInputs) float64 {
	return 0.40*clamp01(in.FleetAcceptanceRate) +
		0.25*clamp01(in.FleetVerificationRate) +
		0.15*(1-clamp01(in.FleetReworkRate)) +
		0.10*(1-clamp01(in.DisputeRate)) +
		0.10*clamp01(in.UptimeScore)
}

func ComputeEffectiveRF(agentRF, operatorRF float64) float64 {
	return 0.65*clamp01(agentRF) + 0.35*clamp01(operatorRF)
}

func ComputeFinalRF(rfLast100, rfLast30d, rfLifetime float64) float64 {
	return 0.50*clamp01(rfLast100) + 0.30*clamp01(rfLast30d) + 0.20*clamp01(rfLifetime)
}

func RewardMultiplierFromRF(rf float64) float64 {
	v := clamp01(rf)
	switch {
	case v >= 0.95:
		return 1.00
	case v >= 0.90:
		return 0.92
	case v >= 0.80:
		return 0.75
	case v >= 0.70:
		return 0.50
	case v >= 0.55:
		return 0.20
	default:
		return 0.00
	}
}

func gpuWeight(class string) float64 {
	switch strings.ToLower(strings.TrimSpace(class)) {
	case "cpu-only", "cpu_only", "cpu":
		return 0.010
	case "mps":
		return 0.018
	case "rtx_3060_class":
		return 0.025
	case "rtx_4090_class":
		return 0.050
	case "a100_h100_class":
		return 0.090
	default:
		return 0.010
	}
}

func taskMultiplier(difficulty string) float64 {
	switch strings.ToLower(strings.TrimSpace(difficulty)) {
	case "easy":
		return 0.8
	case "hard":
		return 1.25
	case "critical":
		return 1.6
	default:
		return 1.0
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
