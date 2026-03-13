package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrApprovalGateRequired = errors.New("approval gate is required")
	ErrInvalidLoopPolicy    = errors.New("invalid loop policy")
)

type IterationCandidate struct {
	CandidateID     string
	BenchmarkDelta  float64
	EstimatedCostJW float64
}

type LoopPolicy struct {
	MaxIterations   int
	BudgetLimitJW   float64
	RequireApproval bool
	ApprovalGranted bool
	MinAcceptScore  float64
}

type LoopRequest struct {
	ExperimentID string
	Candidates   []IterationCandidate
	Policy       LoopPolicy
}

type IterationRecord struct {
	Iteration      int
	CandidateID    string
	Score          float64
	CostJW         float64
	Decision       string
	VerifierReason string
}

type LoopResult struct {
	ExperimentID        string
	Status              string
	StopReason          string
	AcceptedCandidateID string
	TotalCostJW         float64
	Iterations          []IterationRecord
}

type EvaluatorHook interface {
	Evaluate(ctx context.Context, iteration int, candidate IterationCandidate) (score float64, costJW float64, details map[string]any, err error)
}

type VerifierHook interface {
	Verify(ctx context.Context, iteration int, candidate IterationCandidate, score float64) (accepted bool, reason string, details map[string]any, err error)
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

func (s *Service) Run(ctx context.Context, req LoopRequest) (LoopResult, error) {
	policy, err := normalizePolicy(req.Policy)
	if err != nil {
		return LoopResult{}, err
	}
	if policy.RequireApproval && !policy.ApprovalGranted {
		return LoopResult{}, ErrApprovalGateRequired
	}
	if len(req.Candidates) == 0 {
		return LoopResult{
			ExperimentID: strings.TrimSpace(req.ExperimentID),
			Status:       "completed",
			StopReason:   "no_candidates",
		}, nil
	}

	result := LoopResult{
		ExperimentID: strings.TrimSpace(req.ExperimentID),
		Status:       "completed",
		StopReason:   "iteration_limit_reached",
		Iterations:   make([]IterationRecord, 0, policy.MaxIterations),
	}

	maxIterations := policy.MaxIterations
	if len(req.Candidates) < maxIterations {
		maxIterations = len(req.Candidates)
	}

	for i := 0; i < maxIterations; i++ {
		candidate := normalizeCandidate(req.Candidates[i], i)
		score, costJW, _, err := s.evaluator.Evaluate(ctx, i+1, candidate)
		if err != nil {
			return LoopResult{}, fmt.Errorf("evaluate iteration %d: %w", i+1, err)
		}
		if result.TotalCostJW+costJW > policy.BudgetLimitJW {
			result.StopReason = "budget_exceeded"
			result.Status = "halted"
			break
		}

		result.TotalCostJW += costJW
		accepted, reason, _, err := s.verifier.Verify(ctx, i+1, candidate, score)
		if err != nil {
			return LoopResult{}, fmt.Errorf("verify iteration %d: %w", i+1, err)
		}

		decision := "rejected"
		if accepted {
			decision = "accepted"
		} else if score >= policy.MinAcceptScore*0.85 {
			decision = "needs_review"
		}
		result.Iterations = append(result.Iterations, IterationRecord{
			Iteration:      i + 1,
			CandidateID:    candidate.CandidateID,
			Score:          score,
			CostJW:         costJW,
			Decision:       decision,
			VerifierReason: reason,
		})

		if accepted && score >= policy.MinAcceptScore {
			result.StopReason = "accepted_candidate"
			result.AcceptedCandidateID = candidate.CandidateID
			break
		}
	}

	return result, nil
}

func normalizePolicy(in LoopPolicy) (LoopPolicy, error) {
	if in.MaxIterations <= 0 {
		in.MaxIterations = 3
	}
	if in.MaxIterations > 25 {
		return LoopPolicy{}, fmt.Errorf("%w: max_iterations must be <= 25", ErrInvalidLoopPolicy)
	}
	if in.BudgetLimitJW <= 0 {
		in.BudgetLimitJW = 10
	}
	if in.MinAcceptScore <= 0 {
		in.MinAcceptScore = 0.62
	}
	if in.MinAcceptScore > 1 {
		return LoopPolicy{}, fmt.Errorf("%w: min_accept_score must be <= 1", ErrInvalidLoopPolicy)
	}
	return in, nil
}

func normalizeCandidate(in IterationCandidate, index int) IterationCandidate {
	candidateID := strings.TrimSpace(in.CandidateID)
	if candidateID == "" {
		candidateID = fmt.Sprintf("candidate_%d", index+1)
	}
	cost := in.EstimatedCostJW
	if cost <= 0 {
		cost = 1
	}
	return IterationCandidate{
		CandidateID:     candidateID,
		BenchmarkDelta:  in.BenchmarkDelta,
		EstimatedCostJW: cost,
	}
}

type defaultEvaluator struct{}

func (defaultEvaluator) Evaluate(_ context.Context, _ int, candidate IterationCandidate) (float64, float64, map[string]any, error) {
	score := 0.50 + candidate.BenchmarkDelta
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	cost := candidate.EstimatedCostJW
	if cost <= 0 {
		cost = 1
	}
	return score, cost, map[string]any{
		"benchmark_delta": candidate.BenchmarkDelta,
	}, nil
}

type defaultVerifier struct{}

func (defaultVerifier) Verify(_ context.Context, _ int, _ IterationCandidate, score float64) (bool, string, map[string]any, error) {
	if score >= 0.62 {
		return true, "score passed verifier threshold", map[string]any{"threshold": 0.62}, nil
	}
	return false, "score below verifier threshold", map[string]any{"threshold": 0.62}, nil
}
