package simulation

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ScopePlatform = "platform"
	ScopeTenant   = "tenant"
)

const (
	ScenarioStatusDraft     = "draft"
	ScenarioStatusPublished = "published"
)

const (
	RunModeDeterministicStub = "deterministic_stub"
	RunModeSampledSynthetic  = "sampled_synthetic"
	RunModeRuntimeBacked     = "runtime_backed"
)

const (
	RunStatusQueued    = "queued"
	RunStatusRunning   = "running"
	RunStatusCompleted = "completed"
	RunStatusCancelled = "cancelled"
)

const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

const (
	ReplayStatusPendingApproval = "pending_approval"
	ReplayStatusApproved        = "approved"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrInvalidRequest = errors.New("invalid request")
)

type Scenario struct {
	TenantID         string    `json:"tenant_id"`
	ScenarioID       string    `json:"scenario_id"`
	Scope            string    `json:"scope"`
	Name             string    `json:"name"`
	Goal             string    `json:"goal"`
	Status           string    `json:"status"`
	PublishedVersion int       `json:"published_version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ScenarioVersion struct {
	TenantID   string         `json:"tenant_id"`
	ScenarioID string         `json:"scenario_id"`
	Version    int            `json:"version"`
	Spec       map[string]any `json:"spec"`
	CreatedAt  time.Time      `json:"created_at"`
}

type RunEvent struct {
	Type       string         `json:"type"`
	Payload    map[string]any `json:"payload"`
	OccurredAt time.Time      `json:"occurred_at"`
}

type RunMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit,omitempty"`
}

type RunFinding struct {
	FindingType string         `json:"finding_type"`
	Severity    string         `json:"severity"`
	Summary     string         `json:"summary"`
	Metadata    map[string]any `json:"metadata"`
}

type RunArtifact struct {
	ArtifactID string         `json:"artifact_id"`
	Kind       string         `json:"kind"`
	Ref        string         `json:"ref"`
	Metadata   map[string]any `json:"metadata"`
}

type Run struct {
	TenantID        string         `json:"tenant_id"`
	RunID           string         `json:"run_id"`
	ScenarioID      string         `json:"scenario_id"`
	ScenarioVersion int            `json:"scenario_version"`
	RunMode         string         `json:"run_mode"`
	Seed            string         `json:"seed"`
	ScaleProfile    string         `json:"scale_profile"`
	Status          string         `json:"status"`
	BudgetLimitJW   float64        `json:"budget_limit_jw,omitempty"`
	TimeboxSeconds  int            `json:"timebox_seconds,omitempty"`
	PolicyOverrides map[string]any `json:"policy_overrides,omitempty"`
	Events          []RunEvent     `json:"events"`
	Metrics         []RunMetric    `json:"metrics"`
	Findings        []RunFinding   `json:"findings"`
	Artifacts       []RunArtifact  `json:"artifacts"`
	CreatedAt       time.Time      `json:"created_at"`
	CompletedAt     time.Time      `json:"completed_at"`
}

type Baseline struct {
	TenantID        string    `json:"tenant_id"`
	BaselineID      string    `json:"baseline_id"`
	ScenarioID      string    `json:"scenario_id"`
	ScenarioVersion int       `json:"scenario_version"`
	RunID           string    `json:"run_id"`
	Reason          string    `json:"reason"`
	PromotedBy      string    `json:"promoted_by"`
	CreatedAt       time.Time `json:"created_at"`
}

type MetricDiff struct {
	Name          string  `json:"name"`
	Candidate     float64 `json:"candidate"`
	Baseline      float64 `json:"baseline"`
	DeltaAbsolute float64 `json:"delta_absolute"`
}

type CompareResult struct {
	TenantID       string       `json:"tenant_id"`
	CandidateRunID string       `json:"candidate_run_id"`
	BaselineID     string       `json:"baseline_id"`
	Verdict        string       `json:"verdict"`
	MetricDiffs    []MetricDiff `json:"metric_diffs"`
	Reasons        []string     `json:"reasons"`
}

type Replay struct {
	TenantID        string    `json:"tenant_id"`
	ReplayID        string    `json:"replay_id"`
	SourceType      string    `json:"source_type"`
	SourceRef       string    `json:"source_ref"`
	Sensitivity     string    `json:"sensitivity"`
	Status          string    `json:"status"`
	ApprovalGranted bool      `json:"approval_granted"`
	RequestedBy     string    `json:"requested_by"`
	CreatedAt       time.Time `json:"created_at"`
}

type RiskHeatmapCell struct {
	ScenarioID             string  `json:"scenario_id"`
	RiskScore              float64 `json:"risk_score"`
	HighOrAboveFindings    int     `json:"high_or_above_findings"`
	TotalFindings          int     `json:"total_findings"`
	SchedulerStarvation    float64 `json:"scheduler_starvation"`
	FalseAcceptPenetration float64 `json:"false_accept_penetration"`
	CriticMonocultureRatio float64 `json:"critic_monoculture_ratio"`
}

type Dashboard struct {
	TenantID           string            `json:"tenant_id"`
	TotalRuns          int               `json:"total_runs"`
	StatusCounts       map[string]int    `json:"status_counts"`
	RunModeCounts      map[string]int    `json:"run_mode_counts"`
	ScenarioCounts     map[string]int    `json:"scenario_counts"`
	FindingsBySeverity map[string]int    `json:"findings_by_severity"`
	BaselinesTotal     int               `json:"baselines_total"`
	RecentRuns         []Run             `json:"recent_runs"`
	RiskHeatmap        []RiskHeatmapCell `json:"risk_heatmap"`
}

type CreateScenarioRequest struct {
	TenantID   string
	ScenarioID string
	Scope      string
	Name       string
	Goal       string
}

type CreateScenarioVersionRequest struct {
	Version int
	Spec    map[string]any
}

type StartRunRequest struct {
	TenantID        string
	ScenarioID      string
	ScenarioVersion int
	RunMode         string
	Seed            string
	ScaleProfile    string
	BudgetLimitJW   float64
	TimeboxSeconds  int
	PolicyOverrides map[string]any
}

type PromoteBaselineRequest struct {
	TenantID   string
	RunID      string
	Reason     string
	PromotedBy string
}

type CompareRequest struct {
	TenantID         string
	CandidateRunID   string
	BaselineID       string
	FailOnSeverities []string
}

type RequestReplayRequest struct {
	TenantID        string
	SourceType      string
	SourceRef       string
	Sensitivity     string
	RequestedBy     string
	ApprovalGranted bool
}

type Service struct {
	mu sync.RWMutex

	scenarios       map[string]Scenario
	scenarioVersion map[string]ScenarioVersion
	runs            map[string]Run
	baselines       map[string]Baseline
	replays         map[string]Replay
	seededTenants   map[string]bool
	seq             uint64
}

func New() *Service {
	return &Service{
		scenarios:       map[string]Scenario{},
		scenarioVersion: map[string]ScenarioVersion{},
		runs:            map[string]Run{},
		baselines:       map[string]Baseline{},
		replays:         map[string]Replay{},
		seededTenants:   map[string]bool{},
	}
}

type scenarioSeed struct {
	ScenarioID string
	Name       string
	Goal       string
}

var v1ScenarioSeeds = []scenarioSeed{
	{ScenarioID: "scheduler_starvation_v1", Name: "Scheduler Starvation", Goal: "Detect unfair dispatch and starvation."},
	{ScenarioID: "verifier_collusion_v1", Name: "Verifier Collusion", Goal: "Detect verifier agreement inflation and false passes."},
	{ScenarioID: "reducer_instability_v1", Name: "Reducer Instability", Goal: "Measure reduction determinism under reruns."},
	{ScenarioID: "market_spam_attack_v1", Name: "Market Spam Attack", Goal: "Stress pricing and reserve defenses against spam floods."},
	{ScenarioID: "approval_backlog_v1", Name: "Approval Backlog", Goal: "Test queue growth and SLA degradation."},
	{ScenarioID: "research_false_consensus_v1", Name: "Research False Consensus", Goal: "Probe quorum quality under shared bad evidence."},
	{ScenarioID: "code_patch_merge_storm_v1", Name: "Code Patch Merge Storm", Goal: "Stress merge and reduction under conflict storms."},
	{ScenarioID: "autosearch_reward_hacking_v1", Name: "Autosearch Reward Hacking", Goal: "Detect quality gaming in bounded self-improve loops."},
	{ScenarioID: "incident_replay_v1", Name: "Incident Replay", Goal: "Replay archived incidents against candidate mitigations."},
	{ScenarioID: "critic_monoculture_v1", Name: "Critic Monoculture", Goal: "Measure diversity collapse and agreement inflation."},
	{ScenarioID: "acceptance_contract_drift_v1", Name: "Acceptance Contract Drift", Goal: "Detect stale contract false accepts/rejects."},
}

func (s *Service) EnsureV1ScenarioLibrary(_ context.Context, tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.seededTenants[tenantID] {
		return nil
	}

	now := time.Now().UTC()
	for _, seed := range v1ScenarioSeeds {
		scenarioKey := tenantScoped(tenantID, seed.ScenarioID)
		scenario := Scenario{
			TenantID:         tenantID,
			ScenarioID:       seed.ScenarioID,
			Scope:            ScopeTenant,
			Name:             seed.Name,
			Goal:             seed.Goal,
			Status:           ScenarioStatusPublished,
			PublishedVersion: 1,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if existing, ok := s.scenarios[scenarioKey]; ok {
			scenario = existing
			scenario.Status = ScenarioStatusPublished
			if scenario.PublishedVersion <= 0 {
				scenario.PublishedVersion = 1
			}
			scenario.UpdatedAt = now
		}
		s.scenarios[scenarioKey] = scenario

		versionKey := scenarioVersionKey(tenantID, seed.ScenarioID, 1)
		if _, ok := s.scenarioVersion[versionKey]; !ok {
			s.scenarioVersion[versionKey] = ScenarioVersion{
				TenantID:   tenantID,
				ScenarioID: seed.ScenarioID,
				Version:    1,
				Spec: map[string]any{
					"goal":             seed.Goal,
					"seed_strategy":    "fixed",
					"runtime_modes":    []string{RunModeDeterministicStub, RunModeSampledSynthetic},
					"workload_mix":     map[string]any{"research.extract": 0.3, "code.patch": 0.4, "reduce.merge": 0.3},
					"population_model": map[string]any{"workers": 24, "verifiers": 8, "reducers": 4},
				},
				CreatedAt: now,
			}
		}
	}
	s.seededTenants[tenantID] = true
	return nil
}

func (s *Service) CreateScenario(_ context.Context, req CreateScenarioRequest) (Scenario, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ScenarioID) == "" || strings.TrimSpace(req.Name) == "" {
		return Scenario{}, fmt.Errorf("%w: tenant_id, scenario_id, and name are required", ErrInvalidRequest)
	}
	if req.Scope == "" {
		req.Scope = ScopeTenant
	}
	if req.Scope != ScopeTenant && req.Scope != ScopePlatform {
		return Scenario{}, fmt.Errorf("%w: invalid scope", ErrInvalidRequest)
	}

	now := time.Now().UTC()
	key := tenantScoped(req.TenantID, req.ScenarioID)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.scenarios[key]; ok {
		return Scenario{}, fmt.Errorf("%w: scenario_id", ErrAlreadyExists)
	}

	scenario := Scenario{
		TenantID:   req.TenantID,
		ScenarioID: req.ScenarioID,
		Scope:      req.Scope,
		Name:       req.Name,
		Goal:       req.Goal,
		Status:     ScenarioStatusDraft,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	s.scenarios[key] = scenario
	return scenario, nil
}

func (s *Service) ListScenarios(_ context.Context, tenantID string) []Scenario {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Scenario, 0)
	for _, sc := range s.scenarios {
		if sc.TenantID != tenantID {
			continue
		}
		items = append(items, sc)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ScenarioID < items[j].ScenarioID
	})
	return items
}

func (s *Service) GetScenario(_ context.Context, tenantID, scenarioID string) (Scenario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scenario, ok := s.scenarios[tenantScoped(tenantID, scenarioID)]
	if !ok {
		return Scenario{}, ErrNotFound
	}
	return scenario, nil
}

func (s *Service) CreateScenarioVersion(_ context.Context, tenantID, scenarioID string, req CreateScenarioVersionRequest) (ScenarioVersion, error) {
	if req.Version <= 0 {
		return ScenarioVersion{}, fmt.Errorf("%w: version must be > 0", ErrInvalidRequest)
	}
	if req.Spec == nil {
		req.Spec = map[string]any{}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.scenarios[tenantScoped(tenantID, scenarioID)]; !ok {
		return ScenarioVersion{}, ErrNotFound
	}
	k := scenarioVersionKey(tenantID, scenarioID, req.Version)
	if _, ok := s.scenarioVersion[k]; ok {
		return ScenarioVersion{}, fmt.Errorf("%w: scenario version", ErrAlreadyExists)
	}

	version := ScenarioVersion{
		TenantID:   tenantID,
		ScenarioID: scenarioID,
		Version:    req.Version,
		Spec:       copyMap(req.Spec),
		CreatedAt:  time.Now().UTC(),
	}
	s.scenarioVersion[k] = version
	return version, nil
}

func (s *Service) PublishScenario(_ context.Context, tenantID, scenarioID string, version int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := tenantScoped(tenantID, scenarioID)
	scenario, ok := s.scenarios[key]
	if !ok {
		return ErrNotFound
	}
	if _, ok := s.scenarioVersion[scenarioVersionKey(tenantID, scenarioID, version)]; !ok {
		return ErrNotFound
	}
	scenario.Status = ScenarioStatusPublished
	scenario.PublishedVersion = version
	scenario.UpdatedAt = time.Now().UTC()
	s.scenarios[key] = scenario
	return nil
}

func (s *Service) StartRun(_ context.Context, req StartRunRequest) (Run, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ScenarioID) == "" {
		return Run{}, fmt.Errorf("%w: tenant_id and scenario_id are required", ErrInvalidRequest)
	}
	if req.ScenarioVersion <= 0 {
		return Run{}, fmt.Errorf("%w: scenario_version must be > 0", ErrInvalidRequest)
	}
	if req.RunMode == "" {
		req.RunMode = RunModeDeterministicStub
	}
	if !isRunMode(req.RunMode) {
		return Run{}, fmt.Errorf("%w: invalid run_mode", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	scenarioKey := tenantScoped(req.TenantID, req.ScenarioID)
	scenario, ok := s.scenarios[scenarioKey]
	if !ok {
		return Run{}, ErrNotFound
	}
	if _, ok := s.scenarioVersion[scenarioVersionKey(req.TenantID, req.ScenarioID, req.ScenarioVersion)]; !ok {
		return Run{}, ErrNotFound
	}
	if scenario.Status != ScenarioStatusPublished {
		return Run{}, fmt.Errorf("%w: scenario must be published", ErrInvalidRequest)
	}

	now := time.Now().UTC()
	runID := s.nextIDLocked("simrun")
	seed := req.Seed
	if strings.TrimSpace(seed) == "" {
		seed = "seed_default"
	}

	var metrics []RunMetric
	var findings []RunFinding
	var artifacts []RunArtifact
	switch req.RunMode {
	case RunModeDeterministicStub:
		metrics, findings, artifacts = deterministicOutputs(req.ScenarioID, seed)
	case RunModeSampledSynthetic:
		metrics, findings, artifacts = sampledSyntheticOutputs(req.ScenarioID, seed)
	case RunModeRuntimeBacked:
		metrics, findings, artifacts = runtimeBackedOutputs(req.ScenarioID, seed)
	}
	run := Run{
		TenantID:        req.TenantID,
		RunID:           runID,
		ScenarioID:      req.ScenarioID,
		ScenarioVersion: req.ScenarioVersion,
		RunMode:         req.RunMode,
		Seed:            seed,
		ScaleProfile:    req.ScaleProfile,
		Status:          RunStatusCompleted,
		BudgetLimitJW:   req.BudgetLimitJW,
		TimeboxSeconds:  req.TimeboxSeconds,
		PolicyOverrides: copyMap(req.PolicyOverrides),
		Events: []RunEvent{
			{Type: "simulation.run_started", Payload: map[string]any{"run_mode": req.RunMode, "seed": seed}, OccurredAt: now},
			{Type: "simulation.finding_created", Payload: map[string]any{"count": len(findings)}, OccurredAt: now},
			{Type: "simulation.run_completed", Payload: map[string]any{"status": RunStatusCompleted}, OccurredAt: now},
		},
		Metrics:     metrics,
		Findings:    findings,
		Artifacts:   artifacts,
		CreatedAt:   now,
		CompletedAt: now,
	}

	s.runs[tenantScoped(req.TenantID, runID)] = run
	return run, nil
}

func (s *Service) ListRuns(_ context.Context, tenantID, scenarioID string) []Run {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Run, 0)
	for _, run := range s.runs {
		if run.TenantID != tenantID {
			continue
		}
		if strings.TrimSpace(scenarioID) != "" && run.ScenarioID != scenarioID {
			continue
		}
		items = append(items, run)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *Service) GetRun(_ context.Context, tenantID, runID string) (Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	run, ok := s.runs[tenantScoped(tenantID, runID)]
	if !ok {
		return Run{}, ErrNotFound
	}
	return run, nil
}

func (s *Service) CancelRun(_ context.Context, tenantID, runID string) (Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := tenantScoped(tenantID, runID)
	run, ok := s.runs[k]
	if !ok {
		return Run{}, ErrNotFound
	}
	if run.Status == RunStatusCompleted {
		return run, nil
	}
	run.Status = RunStatusCancelled
	s.runs[k] = run
	return run, nil
}

func (s *Service) ListRunEvents(_ context.Context, tenantID, runID string) ([]RunEvent, error) {
	run, err := s.GetRun(context.Background(), tenantID, runID)
	if err != nil {
		return nil, err
	}
	return run.Events, nil
}

func (s *Service) ListRunMetrics(_ context.Context, tenantID, runID string) ([]RunMetric, error) {
	run, err := s.GetRun(context.Background(), tenantID, runID)
	if err != nil {
		return nil, err
	}
	return run.Metrics, nil
}

func (s *Service) ListRunFindings(_ context.Context, tenantID, runID string) ([]RunFinding, error) {
	run, err := s.GetRun(context.Background(), tenantID, runID)
	if err != nil {
		return nil, err
	}
	return run.Findings, nil
}

func (s *Service) ListRunArtifacts(_ context.Context, tenantID, runID string) ([]RunArtifact, error) {
	run, err := s.GetRun(context.Background(), tenantID, runID)
	if err != nil {
		return nil, err
	}
	return run.Artifacts, nil
}

func (s *Service) PromoteBaseline(_ context.Context, req PromoteBaselineRequest) (Baseline, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.RunID) == "" {
		return Baseline{}, fmt.Errorf("%w: tenant_id and run_id are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	run, ok := s.runs[tenantScoped(req.TenantID, req.RunID)]
	if !ok {
		return Baseline{}, ErrNotFound
	}
	if run.Status != RunStatusCompleted {
		return Baseline{}, fmt.Errorf("%w: run must be completed", ErrInvalidRequest)
	}

	baseline := Baseline{
		TenantID:        req.TenantID,
		BaselineID:      s.nextIDLocked("base"),
		ScenarioID:      run.ScenarioID,
		ScenarioVersion: run.ScenarioVersion,
		RunID:           run.RunID,
		Reason:          req.Reason,
		PromotedBy:      req.PromotedBy,
		CreatedAt:       time.Now().UTC(),
	}
	s.baselines[tenantScoped(req.TenantID, baseline.BaselineID)] = baseline
	return baseline, nil
}

func (s *Service) ListBaselines(_ context.Context, tenantID, scenarioID string) []Baseline {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Baseline, 0)
	for _, baseline := range s.baselines {
		if baseline.TenantID != tenantID {
			continue
		}
		if strings.TrimSpace(scenarioID) != "" && baseline.ScenarioID != scenarioID {
			continue
		}
		items = append(items, baseline)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *Service) GetBaseline(_ context.Context, tenantID, baselineID string) (Baseline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	baseline, ok := s.baselines[tenantScoped(tenantID, baselineID)]
	if !ok {
		return Baseline{}, ErrNotFound
	}
	return baseline, nil
}

func (s *Service) Compare(_ context.Context, req CompareRequest) (CompareResult, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.CandidateRunID) == "" || strings.TrimSpace(req.BaselineID) == "" {
		return CompareResult{}, fmt.Errorf("%w: tenant_id, candidate_run_id, and baseline_id are required", ErrInvalidRequest)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	candidate, ok := s.runs[tenantScoped(req.TenantID, req.CandidateRunID)]
	if !ok {
		return CompareResult{}, ErrNotFound
	}
	baseline, ok := s.baselines[tenantScoped(req.TenantID, req.BaselineID)]
	if !ok {
		return CompareResult{}, ErrNotFound
	}
	baselineRun, ok := s.runs[tenantScoped(req.TenantID, baseline.RunID)]
	if !ok {
		return CompareResult{}, ErrNotFound
	}

	baseMetrics := map[string]float64{}
	for _, m := range baselineRun.Metrics {
		baseMetrics[m.Name] = m.Value
	}
	diffs := make([]MetricDiff, 0)
	for _, m := range candidate.Metrics {
		b := baseMetrics[m.Name]
		delta := m.Value - b
		if delta < 0 {
			delta = -delta
		}
		diffs = append(diffs, MetricDiff{Name: m.Name, Candidate: m.Value, Baseline: b, DeltaAbsolute: delta})
	}

	failSeverities := map[string]struct{}{}
	for _, sev := range req.FailOnSeverities {
		failSeverities[strings.ToLower(strings.TrimSpace(sev))] = struct{}{}
	}

	verdict := "pass"
	reasons := make([]string, 0)
	for _, f := range candidate.Findings {
		if _, ok := failSeverities[strings.ToLower(f.Severity)]; ok {
			verdict = "fail"
			reasons = append(reasons, "finding severity triggered fail: "+f.Severity)
		}
	}
	if verdict != "fail" {
		for _, d := range diffs {
			if d.DeltaAbsolute > 0.2 {
				verdict = "review"
				reasons = append(reasons, "metric drift exceeded 0.2: "+d.Name)
				break
			}
		}
	}

	return CompareResult{
		TenantID:       req.TenantID,
		CandidateRunID: req.CandidateRunID,
		BaselineID:     req.BaselineID,
		Verdict:        verdict,
		MetricDiffs:    diffs,
		Reasons:        reasons,
	}, nil
}

func (s *Service) RequestReplay(_ context.Context, req RequestReplayRequest) (Replay, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.SourceType) == "" || strings.TrimSpace(req.SourceRef) == "" {
		return Replay{}, fmt.Errorf("%w: tenant_id, source_type, and source_ref are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	status := ReplayStatusApproved
	if strings.EqualFold(strings.TrimSpace(req.Sensitivity), "sensitive") && !req.ApprovalGranted {
		status = ReplayStatusPendingApproval
	}

	replay := Replay{
		TenantID:        req.TenantID,
		ReplayID:        s.nextIDLocked("replay"),
		SourceType:      req.SourceType,
		SourceRef:       req.SourceRef,
		Sensitivity:     req.Sensitivity,
		Status:          status,
		ApprovalGranted: req.ApprovalGranted,
		RequestedBy:     req.RequestedBy,
		CreatedAt:       time.Now().UTC(),
	}
	s.replays[tenantScoped(req.TenantID, replay.ReplayID)] = replay
	return replay, nil
}

func (s *Service) GetReplay(_ context.Context, tenantID, replayID string) (Replay, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	replay, ok := s.replays[tenantScoped(tenantID, replayID)]
	if !ok {
		return Replay{}, ErrNotFound
	}
	return replay, nil
}

func (s *Service) Dashboard(_ context.Context, tenantID string, limit int) Dashboard {
	if limit <= 0 {
		limit = 20
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	runs := make([]Run, 0)
	statusCounts := map[string]int{}
	runModeCounts := map[string]int{}
	scenarioCounts := map[string]int{}
	findingsBySeverity := map[string]int{}
	riskByScenario := map[string]RiskHeatmapCell{}
	baselinesTotal := 0

	for _, baseline := range s.baselines {
		if baseline.TenantID != tenantID {
			continue
		}
		baselinesTotal++
	}

	for _, run := range s.runs {
		if run.TenantID != tenantID {
			continue
		}
		runs = append(runs, run)
		statusCounts[run.Status]++
		runModeCounts[run.RunMode]++
		scenarioCounts[run.ScenarioID]++

		cell := riskByScenario[run.ScenarioID]
		cell.ScenarioID = run.ScenarioID

		for _, metric := range run.Metrics {
			switch metric.Name {
			case "scheduler_starvation":
				if metric.Value > cell.SchedulerStarvation {
					cell.SchedulerStarvation = metric.Value
				}
			case "false_accept_penetration":
				if metric.Value > cell.FalseAcceptPenetration {
					cell.FalseAcceptPenetration = metric.Value
				}
			case "critic_monoculture_ratio":
				if metric.Value > cell.CriticMonocultureRatio {
					cell.CriticMonocultureRatio = metric.Value
				}
			}
		}

		for _, finding := range run.Findings {
			severity := strings.ToLower(strings.TrimSpace(finding.Severity))
			findingsBySeverity[severity]++
			cell.TotalFindings++
			if severity == SeverityHigh || severity == SeverityCritical {
				cell.HighOrAboveFindings++
			}
			if w := severityWeight(severity); w > cell.RiskScore {
				cell.RiskScore = w
			}
		}

		metricRisk := maxFloat(cell.SchedulerStarvation, cell.FalseAcceptPenetration, cell.CriticMonocultureRatio)
		if metricRisk > cell.RiskScore {
			cell.RiskScore = metricRisk
		}
		riskByScenario[run.ScenarioID] = cell
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].CreatedAt.After(runs[j].CreatedAt)
	})
	recentRuns := runs
	if len(recentRuns) > limit {
		recentRuns = recentRuns[:limit]
	}

	riskHeatmap := make([]RiskHeatmapCell, 0, len(riskByScenario))
	for _, cell := range riskByScenario {
		riskHeatmap = append(riskHeatmap, cell)
	}
	sort.Slice(riskHeatmap, func(i, j int) bool {
		if riskHeatmap[i].RiskScore == riskHeatmap[j].RiskScore {
			return riskHeatmap[i].ScenarioID < riskHeatmap[j].ScenarioID
		}
		return riskHeatmap[i].RiskScore > riskHeatmap[j].RiskScore
	})

	return Dashboard{
		TenantID:           tenantID,
		TotalRuns:          len(runs),
		StatusCounts:       statusCounts,
		RunModeCounts:      runModeCounts,
		ScenarioCounts:     scenarioCounts,
		FindingsBySeverity: findingsBySeverity,
		BaselinesTotal:     baselinesTotal,
		RecentRuns:         recentRuns,
		RiskHeatmap:        riskHeatmap,
	}
}

func isRunMode(v string) bool {
	switch strings.TrimSpace(v) {
	case RunModeDeterministicStub, RunModeSampledSynthetic, RunModeRuntimeBacked:
		return true
	default:
		return false
	}
}

func severityWeight(severity string) float64 {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case SeverityCritical:
		return 1.0
	case SeverityHigh:
		return 0.8
	case SeverityMedium:
		return 0.5
	case SeverityLow:
		return 0.2
	default:
		return 0.1
	}
}

func maxFloat(a float64, rest ...float64) float64 {
	max := a
	for _, v := range rest {
		if v > max {
			max = v
		}
	}
	return max
}

func deterministicOutputs(scenarioID, seed string) ([]RunMetric, []RunFinding, []RunArtifact) {
	h := sha1.Sum([]byte(scenarioID + "::" + seed))
	queueStarvation := boundedFloat(h[0], 0.04, 0.26)
	falseAccept := boundedFloat(h[1], 0.01, 0.18)
	criticMono := boundedFloat(h[2], 0.10, 0.85)

	metrics := []RunMetric{
		{Name: "scheduler_starvation", Value: queueStarvation, Unit: "ratio"},
		{Name: "false_accept_penetration", Value: falseAccept, Unit: "ratio"},
		{Name: "critic_monoculture_ratio", Value: criticMono, Unit: "ratio"},
	}

	severity := SeverityMedium
	if queueStarvation > 0.20 || falseAccept > 0.12 {
		severity = SeverityHigh
	}
	if queueStarvation > 0.24 || falseAccept > 0.16 {
		severity = SeverityCritical
	}

	findings := []RunFinding{
		{
			FindingType: "queue_instability",
			Severity:    severity,
			Summary:     "Queue behavior exceeded soft threshold under synthetic load.",
			Metadata: map[string]any{
				"scheduler_starvation":     queueStarvation,
				"false_accept_penetration": falseAccept,
			},
		},
		{
			FindingType: "critic_monoculture",
			Severity:    SeverityMedium,
			Summary:     "Critic diversity weakened under this seed profile.",
			Metadata: map[string]any{
				"critic_monoculture_ratio": criticMono,
			},
		},
	}

	artifacts := []RunArtifact{
		{
			ArtifactID: "art_sim_" + shortHash(scenarioID+seed),
			Kind:       "report_json",
			Ref:        "s3://simulation/" + scenarioID + "/" + seed + "/report.json",
			Metadata:   map[string]any{"namespace": "simulation"},
		},
	}
	return metrics, findings, artifacts
}

func sampledSyntheticOutputs(scenarioID, seed string) ([]RunMetric, []RunFinding, []RunArtifact) {
	hash := sha1.Sum([]byte("sampled::" + scenarioID + "::" + seed))
	seedInt := int64(binary.BigEndian.Uint64(hash[:8]))
	if seedInt == 0 {
		seedInt = 1
	}
	rng := rand.New(rand.NewSource(seedInt))

	queueStarvation := 0.02 + (rng.Float64() * 0.35)
	falseAccept := 0.005 + (rng.Float64() * 0.22)
	criticMono := 0.08 + (rng.Float64() * 0.88)

	metrics := []RunMetric{
		{Name: "scheduler_starvation", Value: queueStarvation, Unit: "ratio"},
		{Name: "false_accept_penetration", Value: falseAccept, Unit: "ratio"},
		{Name: "critic_monoculture_ratio", Value: criticMono, Unit: "ratio"},
		{Name: "queue_amplification", Value: 1.0 + (rng.Float64() * 4.0), Unit: "factor"},
	}

	severity := SeverityMedium
	if queueStarvation > 0.25 || falseAccept > 0.14 {
		severity = SeverityHigh
	}
	if queueStarvation > 0.30 || falseAccept > 0.18 {
		severity = SeverityCritical
	}

	findings := []RunFinding{
		{
			FindingType: "queue_instability",
			Severity:    severity,
			Summary:     "Sampled synthetic run detected queue drift under stochastic actor behavior.",
			Metadata: map[string]any{
				"seed":                     seed,
				"scheduler_starvation":     queueStarvation,
				"false_accept_penetration": falseAccept,
			},
		},
		{
			FindingType: "critic_monoculture",
			Severity:    SeverityMedium,
			Summary:     "Critic diversity eroded in sampled synthetic actor pool.",
			Metadata: map[string]any{
				"critic_monoculture_ratio": criticMono,
				"seed":                     seed,
			},
		},
	}

	artifacts := []RunArtifact{
		{
			ArtifactID: "art_sim_sampled_" + shortHash(scenarioID+seed),
			Kind:       "sampled_trace_json",
			Ref:        "s3://simulation/" + scenarioID + "/" + seed + "/sampled-trace.json",
			Metadata:   map[string]any{"namespace": "simulation", "mode": RunModeSampledSynthetic},
		},
	}
	return metrics, findings, artifacts
}

func runtimeBackedOutputs(scenarioID, seed string) ([]RunMetric, []RunFinding, []RunArtifact) {
	metrics, findings, artifacts := deterministicOutputs(scenarioID, seed)
	artifacts = append(artifacts, RunArtifact{
		ArtifactID: "art_runtime_" + shortHash(seed+scenarioID),
		Kind:       "runtime_trace_json",
		Ref:        "s3://simulation/" + scenarioID + "/" + seed + "/runtime-trace.json",
		Metadata:   map[string]any{"namespace": "simulation", "mode": RunModeRuntimeBacked},
	})
	return metrics, findings, artifacts
}

func boundedFloat(b byte, low, high float64) float64 {
	ratio := float64(b) / 255.0
	return low + ((high - low) * ratio)
}

func shortHash(v string) string {
	h := sha1.Sum([]byte(v))
	return fmt.Sprintf("%x", h[:4])
}

func tenantScoped(tenantID, id string) string {
	return strings.TrimSpace(tenantID) + "::" + strings.TrimSpace(id)
}

func scenarioVersionKey(tenantID, scenarioID string, version int) string {
	return fmt.Sprintf("%s::%s::%d", strings.TrimSpace(tenantID), strings.TrimSpace(scenarioID), version)
}

func copyMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	buf, _ := json.Marshal(in)
	out := map[string]any{}
	_ = json.Unmarshal(buf, &out)
	return out
}

func (s *Service) nextIDLocked(prefix string) string {
	s.seq++
	return fmt.Sprintf("%s_%06d", prefix, s.seq)
}
