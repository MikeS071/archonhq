package scheduler

import (
	"context"
	"errors"
	"testing"
)

func TestLoopApprovalGate(t *testing.T) {
	svc := New(nil, nil)
	_, err := svc.Run(context.Background(), LoopRequest{
		ExperimentID: "exp_01",
		Candidates: []IterationCandidate{
			{CandidateID: "c1", BenchmarkDelta: 0.2, EstimatedCostJW: 1},
		},
		Policy: LoopPolicy{
			MaxIterations:   3,
			BudgetLimitJW:   5,
			RequireApproval: true,
			ApprovalGranted: false,
		},
	})
	if !errors.Is(err, ErrApprovalGateRequired) {
		t.Fatalf("expected ErrApprovalGateRequired got %v", err)
	}
}

func TestLoopStopsOnBudgetExceeded(t *testing.T) {
	svc := New(nil, nil)
	result, err := svc.Run(context.Background(), LoopRequest{
		ExperimentID: "exp_02",
		Candidates: []IterationCandidate{
			{CandidateID: "c1", BenchmarkDelta: 0.01, EstimatedCostJW: 1.2},
			{CandidateID: "c2", BenchmarkDelta: 0.20, EstimatedCostJW: 1.2},
		},
		Policy: LoopPolicy{
			MaxIterations: 2,
			BudgetLimitJW: 1.5,
		},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if result.StopReason != "budget_exceeded" {
		t.Fatalf("stop reason=%q want budget_exceeded", result.StopReason)
	}
	if len(result.Iterations) != 1 {
		t.Fatalf("iterations=%d want 1", len(result.Iterations))
	}
}

func TestLoopAcceptsCandidate(t *testing.T) {
	svc := New(nil, nil)
	result, err := svc.Run(context.Background(), LoopRequest{
		ExperimentID: "exp_03",
		Candidates: []IterationCandidate{
			{CandidateID: "c1", BenchmarkDelta: 0.13, EstimatedCostJW: 1},
			{CandidateID: "c2", BenchmarkDelta: 0.03, EstimatedCostJW: 1},
		},
		Policy: LoopPolicy{
			MaxIterations: 3,
			BudgetLimitJW: 10,
		},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if result.StopReason != "accepted_candidate" {
		t.Fatalf("stop reason=%q want accepted_candidate", result.StopReason)
	}
	if result.AcceptedCandidateID != "c1" {
		t.Fatalf("accepted candidate=%q want c1", result.AcceptedCandidateID)
	}
}

func TestLoopPolicyValidation(t *testing.T) {
	svc := New(nil, nil)
	_, err := svc.Run(context.Background(), LoopRequest{
		ExperimentID: "exp_04",
		Candidates: []IterationCandidate{
			{CandidateID: "c1", BenchmarkDelta: 0.2, EstimatedCostJW: 1},
		},
		Policy: LoopPolicy{
			MaxIterations: 26,
		},
	})
	if !errors.Is(err, ErrInvalidLoopPolicy) {
		t.Fatalf("expected ErrInvalidLoopPolicy got %v", err)
	}
}
