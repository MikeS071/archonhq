package verification

import (
	"context"
	"testing"
)

func TestEvaluateWithInputScoreOverride(t *testing.T) {
	svc := New(nil, nil)
	score := 0.91
	result, err := svc.Evaluate(context.Background(), Request{
		VerifierType: "compile",
		InputScore:   &score,
	})
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	if result.Score != 0.91 {
		t.Fatalf("score=%v want 0.91", result.Score)
	}
	if result.Decision != "accepted" {
		t.Fatalf("decision=%q want accepted", result.Decision)
	}
}

func TestEvaluateIterativeFormula(t *testing.T) {
	svc := New(nil, nil)
	result, err := svc.Evaluate(context.Background(), Request{
		VerifierType: "benchmark/eval",
		Iterative: &IterativeContext{
			ExperimentID:        "exp_01",
			Iteration:           2,
			CandidateID:         "cand_2",
			BenchmarkDeltaNorm:  0.8,
			EvalReproducibility: 0.9,
			RollbackSafety:      0.7,
			SearchNovelty:       0.5,
			ComputeEfficiency:   0.6,
		},
	})
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	if result.Score <= 0 {
		t.Fatalf("expected positive score")
	}
	if _, ok := result.HookOutputs["iterative"]; !ok {
		t.Fatalf("expected iterative hook output")
	}
	if result.Decision == "" {
		t.Fatalf("expected decision")
	}
}

func TestEvaluateDefaultFallbackAndNeedsReview(t *testing.T) {
	svc := New(nil, nil)
	result, err := svc.Evaluate(context.Background(), Request{
		VerifierType: "compile",
		Report:       map[string]any{},
	})
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	if result.Score != 0.60 {
		t.Fatalf("score=%v want 0.60", result.Score)
	}
	if result.Decision != "needs_review" {
		t.Fatalf("decision=%q want needs_review", result.Decision)
	}
}
