package reduction

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	StrategyAppendOnly   = "append_only_v1"
	StrategyKeyUpsert    = "key_upsert_v1"
	StrategySectionPatch = "section_patch_v1"
	StrategyASTPatch     = "ast_patch_v1"
	StrategyTopKRank     = "topk_rank_v1"
	StrategyQuorumFact   = "quorum_fact_v1"
	StrategyReduceTree   = "reduce_tree_v1"
)

var supportedMergeStrategies = map[string]struct{}{
	StrategyAppendOnly:   {},
	StrategyKeyUpsert:    {},
	StrategySectionPatch: {},
	StrategyASTPatch:     {},
	StrategyTopKRank:     {},
	StrategyQuorumFact:   {},
	StrategyReduceTree:   {},
}

var (
	ErrUnsupportedMergeStrategy = errors.New("unsupported merge strategy")
	ErrNoCandidates             = errors.New("at least one candidate is required")
)

type Candidate struct {
	ResultID string
	Score    float64
	PatchOps []string
	StateRef string
	Metadata map[string]any
}

type MergeRequest struct {
	TaskFamily string
	Strategy   string
	Candidates []Candidate
}

type MergeDecision struct {
	Strategy        string
	Status          string
	WinnerResultID  string
	RankedResultIDs []string
	MergedPatchOps  []string
	OutputStateRef  string
	Explanation     string
}

type Service struct{}

func New() *Service {
	return &Service{}
}

func ResolveStrategy(taskFamily, requested string) string {
	requested = strings.TrimSpace(strings.ToLower(requested))
	if _, ok := supportedMergeStrategies[requested]; ok {
		return requested
	}

	switch strings.TrimSpace(strings.ToLower(taskFamily)) {
	case "code.patch":
		return StrategyASTPatch
	case "doc.section.write":
		return StrategySectionPatch
	case "research.extract":
		return StrategyQuorumFact
	case "reduce.merge":
		return StrategyReduceTree
	case "autosearch.self_improve":
		return StrategyTopKRank
	default:
		return StrategyAppendOnly
	}
}

func (s *Service) Merge(req MergeRequest) (MergeDecision, error) {
	if len(req.Candidates) == 0 {
		return MergeDecision{}, ErrNoCandidates
	}

	strategy := strings.TrimSpace(strings.ToLower(req.Strategy))
	if strategy == "" {
		strategy = ResolveStrategy(req.TaskFamily, "")
	} else if _, ok := supportedMergeStrategies[strategy]; !ok {
		return MergeDecision{}, fmt.Errorf("%w: %s", ErrUnsupportedMergeStrategy, strategy)
	}

	candidates := normalizeCandidates(req.Candidates)
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].ResultID < candidates[j].ResultID
		}
		return candidates[i].Score > candidates[j].Score
	})

	ranked := make([]string, 0, len(candidates))
	for _, c := range candidates {
		ranked = append(ranked, c.ResultID)
	}

	top := candidates[0]
	outputStateRef := strings.TrimSpace(top.StateRef)
	if outputStateRef == "" {
		outputStateRef = "state_" + strings.TrimPrefix(top.ResultID, "res_")
	}

	decision := MergeDecision{
		Strategy:        strategy,
		Status:          "accepted",
		WinnerResultID:  top.ResultID,
		RankedResultIDs: ranked,
		OutputStateRef:  outputStateRef,
	}

	switch strategy {
	case StrategyASTPatch, StrategySectionPatch:
		decision.MergedPatchOps = dedupePatchOps(candidates)
		if top.Score < 0.55 {
			decision.Status = "needs_review"
		}
		decision.Explanation = "Patch merge selected highest-scoring candidate and deduped operations."
	case StrategyReduceTree:
		mergeCandidates := candidates
		if len(mergeCandidates) > 2 {
			mergeCandidates = mergeCandidates[:2]
		}
		decision.MergedPatchOps = dedupePatchOps(mergeCandidates)
		if top.Score < 0.60 {
			decision.Status = "needs_review"
		}
		decision.Explanation = "Reduce-tree merge combined top candidates by score."
	case StrategyTopKRank:
		decision.MergedPatchOps = nil
		if top.Score < 0.65 {
			decision.Status = "needs_review"
		}
		decision.Explanation = "Top-k ranking completed for iterative workload selection."
	case StrategyQuorumFact:
		decision.MergedPatchOps = nil
		if len(candidates) < 2 || top.Score < 0.50 {
			decision.Status = "needs_review"
		}
		decision.Explanation = "Quorum merge evaluated top candidate confidence and cohort depth."
	case StrategyKeyUpsert, StrategyAppendOnly:
		decision.MergedPatchOps = dedupePatchOps(candidates)
		if top.Score < 0.45 {
			decision.Status = "needs_review"
		}
		decision.Explanation = "Deterministic merge applied append/upsert semantics."
	default:
		return MergeDecision{}, fmt.Errorf("%w: %s", ErrUnsupportedMergeStrategy, strategy)
	}

	return decision, nil
}

func normalizeCandidates(in []Candidate) []Candidate {
	out := make([]Candidate, 0, len(in))
	for i, c := range in {
		resultID := strings.TrimSpace(c.ResultID)
		if resultID == "" {
			resultID = fmt.Sprintf("res_candidate_%d", i+1)
		}
		out = append(out, Candidate{
			ResultID: resultID,
			Score:    clamp(c.Score, 0, 1),
			PatchOps: c.PatchOps,
			StateRef: c.StateRef,
			Metadata: c.Metadata,
		})
	}
	return out
}

func dedupePatchOps(candidates []Candidate) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, c := range candidates {
		for _, op := range c.PatchOps {
			op = strings.TrimSpace(op)
			if op == "" {
				continue
			}
			if _, ok := seen[op]; ok {
				continue
			}
			seen[op] = struct{}{}
			out = append(out, op)
		}
	}
	return out
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
