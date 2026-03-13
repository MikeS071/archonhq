package reduction

import (
	"errors"
	"testing"
)

func TestResolveStrategyDefaults(t *testing.T) {
	tests := []struct {
		family string
		want   string
	}{
		{family: "code.patch", want: StrategyASTPatch},
		{family: "doc.section.write", want: StrategySectionPatch},
		{family: "research.extract", want: StrategyQuorumFact},
		{family: "reduce.merge", want: StrategyReduceTree},
		{family: "autosearch.self_improve", want: StrategyTopKRank},
		{family: "unknown", want: StrategyAppendOnly},
	}
	for _, tt := range tests {
		if got := ResolveStrategy(tt.family, ""); got != tt.want {
			t.Fatalf("ResolveStrategy(%q)=%q want %q", tt.family, got, tt.want)
		}
	}
}

func TestMergeASTPatchWinnerAndOps(t *testing.T) {
	svc := New()
	decision, err := svc.Merge(MergeRequest{
		TaskFamily: "code.patch",
		Strategy:   StrategyASTPatch,
		Candidates: []Candidate{
			{ResultID: "res_02", Score: 0.71, PatchOps: []string{"replace:A", "add:C"}, StateRef: "state_02"},
			{ResultID: "res_01", Score: 0.81, PatchOps: []string{"replace:A", "delete:B"}, StateRef: "state_01"},
		},
	})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if decision.WinnerResultID != "res_01" {
		t.Fatalf("winner=%q want res_01", decision.WinnerResultID)
	}
	if decision.Status != "accepted" {
		t.Fatalf("status=%q want accepted", decision.Status)
	}
	if len(decision.MergedPatchOps) != 3 {
		t.Fatalf("expected 3 deduped patch ops got %d", len(decision.MergedPatchOps))
	}
}

func TestMergeGuardrailErrors(t *testing.T) {
	svc := New()

	_, err := svc.Merge(MergeRequest{TaskFamily: "code.patch"})
	if !errors.Is(err, ErrNoCandidates) {
		t.Fatalf("expected ErrNoCandidates got %v", err)
	}

	_, err = svc.Merge(MergeRequest{
		TaskFamily: "code.patch",
		Strategy:   "bad_strategy",
		Candidates: []Candidate{{ResultID: "res_01", Score: 0.7}},
	})
	if !errors.Is(err, ErrUnsupportedMergeStrategy) {
		t.Fatalf("expected ErrUnsupportedMergeStrategy got %v", err)
	}
}

func TestMergeNeedsReviewForLowScore(t *testing.T) {
	svc := New()
	decision, err := svc.Merge(MergeRequest{
		TaskFamily: "reduce.merge",
		Strategy:   StrategyReduceTree,
		Candidates: []Candidate{
			{ResultID: "res_01", Score: 0.40, PatchOps: []string{"x"}},
			{ResultID: "res_02", Score: 0.39, PatchOps: []string{"y"}},
		},
	})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if decision.Status != "needs_review" {
		t.Fatalf("status=%q want needs_review", decision.Status)
	}
}

func TestMergeStrategyCoverage(t *testing.T) {
	svc := New()
	tests := []struct {
		name     string
		strategy string
	}{
		{name: "append-only", strategy: StrategyAppendOnly},
		{name: "key-upsert", strategy: StrategyKeyUpsert},
		{name: "section-patch", strategy: StrategySectionPatch},
		{name: "topk", strategy: StrategyTopKRank},
		{name: "quorum", strategy: StrategyQuorumFact},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := svc.Merge(MergeRequest{
				TaskFamily: "research.extract",
				Strategy:   tt.strategy,
				Candidates: []Candidate{
					{ResultID: "", Score: 1.2, PatchOps: []string{"a", "a", "b"}, StateRef: ""},
					{ResultID: "res_02", Score: -0.2, PatchOps: []string{"c"}, StateRef: "state_02"},
				},
			})
			if err != nil {
				t.Fatalf("merge failed: %v", err)
			}
			if decision.Strategy != tt.strategy {
				t.Fatalf("strategy=%q want %q", decision.Strategy, tt.strategy)
			}
			if len(decision.RankedResultIDs) == 0 {
				t.Fatalf("expected ranked ids")
			}
			if decision.WinnerResultID == "" {
				t.Fatalf("expected winner")
			}
		})
	}
}
