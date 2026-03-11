package scoring

type Metering struct {
	CPUSec            float64 `json:"cpu_sec"`
	GPUSec            float64 `json:"gpu_sec"`
	GPUClass          string  `json:"gpu_class"`
	TokensIn          int64   `json:"tokens_in"`
	TokensOut         int64   `json:"tokens_out"`
	ExternalToolCalls int64   `json:"external_tool_calls"`
	NetworkMB         float64 `json:"network_mb"`
	StorageMB         float64 `json:"storage_mb"`
}

type QualityInputs struct {
	Validity         float64 `json:"validity"`
	VerifierScore    float64 `json:"verifier_score"`
	AcceptanceSignal float64 `json:"acceptance_signal"`
	Novelty          float64 `json:"novelty"`
	LatencyScore     float64 `json:"latency_score"`
}
