package verification

import (
	"context"
	"strings"
)

type IterativeContext struct {
	ExperimentID        string  `json:"experiment_id,omitempty"`
	Iteration           int     `json:"iteration,omitempty"`
	CandidateID         string  `json:"candidate_id,omitempty"`
	BenchmarkDeltaNorm  float64 `json:"benchmark_delta_norm,omitempty"`
	EvalReproducibility float64 `json:"eval_reproducibility,omitempty"`
	RollbackSafety      float64 `json:"rollback_safety,omitempty"`
	SearchNovelty       float64 `json:"search_novelty,omitempty"`
	ComputeEfficiency   float64 `json:"compute_efficiency,omitempty"`
}

type Request struct {
	VerifierType string
	Report       map[string]any
	InputScore   *float64
	Iterative    *IterativeContext
}

type Result struct {
	Score       float64
	Decision    string
	HookOutputs map[string]any
}

type EvaluatorHook interface {
	Evaluate(ctx context.Context, req Request) (score float64, detail map[string]any, err error)
}

type VerifierHook interface {
	Verify(ctx context.Context, req Request, score float64) (decision string, detail map[string]any, err error)
}

type Service struct {
	evaluator EvaluatorHook
	verifier  VerifierHook
}

func New(evaluator EvaluatorHook, verifier VerifierHook) *Service {
	if evaluator == nil {
		evaluator = defaultEvaluator{}
	}
	if verifier == nil {
		verifier = defaultVerifier{}
	}
	return &Service{
		evaluator: evaluator,
		verifier:  verifier,
	}
}

func (s *Service) Evaluate(ctx context.Context, req Request) (Result, error) {
	score, evalDetail, err := s.evaluator.Evaluate(ctx, req)
	if err != nil {
		return Result{}, err
	}

	if req.InputScore != nil {
		score = clamp(*req.InputScore, 0, 1)
	}

	decision, verifierDetail, err := s.verifier.Verify(ctx, req, score)
	if err != nil {
		return Result{}, err
	}

	hookOutputs := map[string]any{
		"evaluator": evalDetail,
		"verifier":  verifierDetail,
	}
	if req.Iterative != nil {
		hookOutputs["iterative"] = map[string]any{
			"experiment_id": req.Iterative.ExperimentID,
			"iteration":     req.Iterative.Iteration,
			"candidate_id":  req.Iterative.CandidateID,
		}
	}

	return Result{
		Score:       score,
		Decision:    decision,
		HookOutputs: hookOutputs,
	}, nil
}

type defaultEvaluator struct{}

func (defaultEvaluator) Evaluate(_ context.Context, req Request) (float64, map[string]any, error) {
	verifierType := strings.TrimSpace(strings.ToLower(req.VerifierType))

	if req.Iterative != nil {
		// Q_autosearch = 0.30*benchmark + 0.25*repro + 0.20*rollback + 0.15*novelty + 0.10*efficiency
		score := 0.30*clamp(req.Iterative.BenchmarkDeltaNorm, 0, 1) +
			0.25*clamp(req.Iterative.EvalReproducibility, 0, 1) +
			0.20*clamp(req.Iterative.RollbackSafety, 0, 1) +
			0.15*clamp(req.Iterative.SearchNovelty, 0, 1) +
			0.10*clamp(req.Iterative.ComputeEfficiency, 0, 1)
		return score, map[string]any{
			"mode":    "iterative",
			"formula": "Q_autosearch",
		}, nil
	}

	if raw, ok := req.Report["score"]; ok {
		switch v := raw.(type) {
		case float64:
			return clamp(v, 0, 1), map[string]any{"mode": "report_score"}, nil
		case int:
			return clamp(float64(v), 0, 1), map[string]any{"mode": "report_score"}, nil
		}
	}

	defaultScore := 0.60
	if verifierType == "policy" || verifierType == "security" {
		defaultScore = 0.55
	}
	return defaultScore, map[string]any{"mode": "default", "verifier_type": verifierType}, nil
}

type defaultVerifier struct{}

func (defaultVerifier) Verify(_ context.Context, req Request, score float64) (string, map[string]any, error) {
	threshold := 0.62
	verifierType := strings.TrimSpace(strings.ToLower(req.VerifierType))
	switch verifierType {
	case "compile", "tests":
		threshold = 0.70
	case "benchmark", "benchmark/eval":
		threshold = 0.66
	case "schema":
		threshold = 0.60
	}

	decision := "rejected"
	if score >= threshold {
		decision = "accepted"
	} else if score >= threshold*0.85 {
		decision = "needs_review"
	}

	return decision, map[string]any{
		"threshold": threshold,
	}, nil
}

func clamp(v, low, high float64) float64 {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
