package assurance

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ValidationTierFast          = "fast"
	ValidationTierStandard      = "standard"
	ValidationTierHighAssurance = "high_assurance"
)

const (
	ContractSourceInline                  = "inline"
	ContractSourceTemplateRef             = "template_ref"
	ContractSourceTemplateRefWithOverride = "template_ref_with_overrides"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

const (
	StagePlan      = "plan"
	StageExecution = "execution"
	StageArtifact  = "artifact"
	StageOutput    = "output"
	StagePolicy    = "policy"
	StageSecurity  = "security"
	StageBenchmark = "benchmark"
	StageReduction = "reduction"
)

const (
	DecisionAccepted    = "accepted"
	DecisionRejected    = "rejected"
	DecisionNeedsReview = "needs_review"
)

const (
	RunStatusCompleted   = "completed"
	RunStatusNeedsReview = "needs_review"
	RunStatusRejected    = "rejected"
	RunStatusEscalated   = "escalated"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrCriticDiversity  = errors.New("critic diversity requirements not met")
	ErrTemplateNotReady = errors.New("template not published")
)

type AcceptanceTemplate struct {
	TenantID         string    `json:"tenant_id"`
	TemplateID       string    `json:"template_id"`
	TaskFamily       string    `json:"task_family"`
	Name             string    `json:"name"`
	Status           string    `json:"status"`
	PublishedVersion int       `json:"published_version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AcceptanceTemplateVersion struct {
	TenantID   string         `json:"tenant_id"`
	TemplateID string         `json:"template_id"`
	Version    int            `json:"version"`
	Contract   map[string]any `json:"contract"`
	CreatedAt  time.Time      `json:"created_at"`
}

type TaskAcceptanceContract struct {
	TenantID         string         `json:"tenant_id"`
	TaskID           string         `json:"task_id"`
	ContractSource   string         `json:"contract_source"`
	TemplateID       string         `json:"template_id,omitempty"`
	TemplateVersion  int            `json:"template_version,omitempty"`
	ValidationTier   string         `json:"validation_tier"`
	ContractSnapshot map[string]any `json:"contract_snapshot"`
	CreatedAt        time.Time      `json:"created_at"`
}

type Critic struct {
	TenantID         string    `json:"tenant_id"`
	CriticID         string    `json:"critic_id"`
	Name             string    `json:"name"`
	StageName        string    `json:"stage_name"`
	TaskFamily       string    `json:"task_family"`
	CriticClass      string    `json:"critic_class"`
	ProviderFamily   string    `json:"provider_family"`
	FailureModeClass string    `json:"failure_mode_class"`
	Enabled          bool      `json:"enabled"`
	Status           string    `json:"status"`
	PublishedVersion int       `json:"published_version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CriticVersion struct {
	TenantID  string         `json:"tenant_id"`
	CriticID  string         `json:"critic_id"`
	Version   int            `json:"version"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"created_at"`
}

type ValidationStageResult struct {
	StageOrder   int       `json:"stage_order"`
	StageName    string    `json:"stage_name"`
	CriticID     string    `json:"critic_id,omitempty"`
	CriticClass  string    `json:"critic_class,omitempty"`
	Decision     string    `json:"decision"`
	Score        float64   `json:"score"`
	EvidenceRefs []string  `json:"evidence_refs"`
	OccurredAt   time.Time `json:"occurred_at"`
}

type ValidationRun struct {
	ValidationRunID string                  `json:"validation_run_id"`
	TenantID        string                  `json:"tenant_id"`
	TaskID          string                  `json:"task_id"`
	TaskFamily      string                  `json:"task_family"`
	ResultReference string                  `json:"result_reference,omitempty"`
	ValidationTier  string                  `json:"validation_tier"`
	Status          string                  `json:"status"`
	Decision        string                  `json:"decision"`
	Contract        TaskAcceptanceContract  `json:"acceptance_contract"`
	StageResults    []ValidationStageResult `json:"stage_results"`
	CreatedAt       time.Time               `json:"created_at"`
	CompletedAt     time.Time               `json:"completed_at"`
}

type ValidationEscalation struct {
	TenantID         string    `json:"tenant_id"`
	ValidationRunID  string    `json:"validation_run_id"`
	Reason           string    `json:"reason"`
	EscalatedByActor string    `json:"escalated_by_actor"`
	CreatedAt        time.Time `json:"created_at"`
}

type ValidationDashboard struct {
	TenantID               string                 `json:"tenant_id"`
	TotalRuns              int                    `json:"total_runs"`
	StatusCounts           map[string]int         `json:"status_counts"`
	DecisionCounts         map[string]int         `json:"decision_counts"`
	TierCounts             map[string]int         `json:"tier_counts"`
	StageIssueCounts       map[string]int         `json:"stage_issue_counts"`
	RecentRuns             []ValidationRun        `json:"recent_runs"`
	EscalationQueue        []ValidationEscalation `json:"escalation_queue"`
	EscalationResidualRate float64                `json:"escalation_residual_rate"`
}

type CreateTemplateRequest struct {
	TenantID   string
	TemplateID string
	TaskFamily string
	Name       string
}

type CreateTemplateVersionRequest struct {
	Version  int
	Contract map[string]any
}

type AttachTaskContractRequest struct {
	TenantID       string
	TaskID         string
	TemplateID     string
	TemplateVer    int
	ValidationTier string
	InlineContract map[string]any
}

type CreateCriticRequest struct {
	TenantID         string
	CriticID         string
	Name             string
	StageName        string
	TaskFamily       string
	CriticClass      string
	ProviderFamily   string
	FailureModeClass string
	Enabled          *bool
}

type CreateCriticVersionRequest struct {
	Version int
	Config  map[string]any
}

type StartValidationRunRequest struct {
	TenantID        string
	TaskID          string
	TaskFamily      string
	ValidationTier  string
	TemplateID      string
	TemplateVersion int
	InlineContract  map[string]any
	ResultReference string
}

type EscalateValidationRunRequest struct {
	TenantID         string
	ValidationRunID  string
	Reason           string
	EscalatedByActor string
}

type Service struct {
	mu sync.RWMutex

	templates        map[string]AcceptanceTemplate
	templateVersions map[string]AcceptanceTemplateVersion
	taskContracts    map[string]TaskAcceptanceContract
	critics          map[string]Critic
	criticVersions   map[string]CriticVersion
	validationRuns   map[string]ValidationRun
	escalations      map[string][]ValidationEscalation
	seq              uint64
}

func New() *Service {
	return &Service{
		templates:        map[string]AcceptanceTemplate{},
		templateVersions: map[string]AcceptanceTemplateVersion{},
		taskContracts:    map[string]TaskAcceptanceContract{},
		critics:          map[string]Critic{},
		criticVersions:   map[string]CriticVersion{},
		validationRuns:   map[string]ValidationRun{},
		escalations:      map[string][]ValidationEscalation{},
	}
}

func (s *Service) CreateTemplate(_ context.Context, req CreateTemplateRequest) (AcceptanceTemplate, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.TemplateID) == "" || strings.TrimSpace(req.TaskFamily) == "" || strings.TrimSpace(req.Name) == "" {
		return AcceptanceTemplate{}, fmt.Errorf("%w: tenant_id, template_id, task_family, and name are required", ErrInvalidRequest)
	}

	now := time.Now().UTC()
	k := tenantScoped(req.TenantID, req.TemplateID)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.templates[k]; ok {
		return AcceptanceTemplate{}, fmt.Errorf("%w: template_id", ErrAlreadyExists)
	}

	tpl := AcceptanceTemplate{
		TenantID:   strings.TrimSpace(req.TenantID),
		TemplateID: strings.TrimSpace(req.TemplateID),
		TaskFamily: strings.TrimSpace(req.TaskFamily),
		Name:       strings.TrimSpace(req.Name),
		Status:     StatusDraft,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	s.templates[k] = tpl
	return tpl, nil
}

func (s *Service) ListTemplates(_ context.Context, tenantID string) []AcceptanceTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]AcceptanceTemplate, 0)
	for _, tpl := range s.templates {
		if tpl.TenantID != tenantID {
			continue
		}
		items = append(items, tpl)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].TemplateID < items[j].TemplateID
	})
	return items
}

func (s *Service) GetTemplate(_ context.Context, tenantID, templateID string) (AcceptanceTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tpl, ok := s.templates[tenantScoped(tenantID, templateID)]
	if !ok {
		return AcceptanceTemplate{}, ErrNotFound
	}
	return tpl, nil
}

func (s *Service) CreateTemplateVersion(_ context.Context, tenantID, templateID string, req CreateTemplateVersionRequest) (AcceptanceTemplateVersion, error) {
	if req.Version <= 0 {
		return AcceptanceTemplateVersion{}, fmt.Errorf("%w: version must be > 0", ErrInvalidRequest)
	}
	if req.Contract == nil {
		return AcceptanceTemplateVersion{}, fmt.Errorf("%w: contract is required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tpl, ok := s.templates[tenantScoped(tenantID, templateID)]
	if !ok {
		return AcceptanceTemplateVersion{}, ErrNotFound
	}

	vk := templateVersionKey(tenantID, templateID, req.Version)
	if _, ok := s.templateVersions[vk]; ok {
		return AcceptanceTemplateVersion{}, fmt.Errorf("%w: template version", ErrAlreadyExists)
	}

	ver := AcceptanceTemplateVersion{
		TenantID:   tenantID,
		TemplateID: templateID,
		Version:    req.Version,
		Contract:   copyMap(req.Contract),
		CreatedAt:  time.Now().UTC(),
	}
	s.templateVersions[vk] = ver
	tpl.UpdatedAt = time.Now().UTC()
	s.templates[tenantScoped(tenantID, templateID)] = tpl

	return ver, nil
}

func (s *Service) PublishTemplate(_ context.Context, tenantID, templateID string, version int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tpl, ok := s.templates[tenantScoped(tenantID, templateID)]
	if !ok {
		return ErrNotFound
	}
	if _, ok := s.templateVersions[templateVersionKey(tenantID, templateID, version)]; !ok {
		return ErrNotFound
	}
	tpl.Status = StatusPublished
	tpl.PublishedVersion = version
	tpl.UpdatedAt = time.Now().UTC()
	s.templates[tenantScoped(tenantID, templateID)] = tpl
	return nil
}

func (s *Service) AttachTaskContract(_ context.Context, req AttachTaskContractRequest) (TaskAcceptanceContract, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.TaskID) == "" {
		return TaskAcceptanceContract{}, fmt.Errorf("%w: tenant_id and task_id are required", ErrInvalidRequest)
	}

	if req.ValidationTier == "" {
		req.ValidationTier = ValidationTierStandard
	}
	if !isValidationTier(req.ValidationTier) {
		return TaskAcceptanceContract{}, fmt.Errorf("%w: invalid validation_tier", ErrInvalidRequest)
	}

	taskKey := tenantScoped(req.TenantID, req.TaskID)

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.taskContracts[taskKey]; ok {
		return existing, nil
	}

	snap := TaskAcceptanceContract{
		TenantID:       strings.TrimSpace(req.TenantID),
		TaskID:         strings.TrimSpace(req.TaskID),
		ValidationTier: strings.TrimSpace(req.ValidationTier),
		CreatedAt:      time.Now().UTC(),
	}

	source := ContractSourceInline
	contract := copyMap(req.InlineContract)

	if strings.TrimSpace(req.TemplateID) != "" {
		tpl, ok := s.templates[tenantScoped(req.TenantID, req.TemplateID)]
		if !ok {
			return TaskAcceptanceContract{}, ErrNotFound
		}
		version := req.TemplateVer
		if version <= 0 {
			version = tpl.PublishedVersion
		}
		if version <= 0 {
			return TaskAcceptanceContract{}, ErrTemplateNotReady
		}
		ver, ok := s.templateVersions[templateVersionKey(req.TenantID, req.TemplateID, version)]
		if !ok {
			return TaskAcceptanceContract{}, ErrNotFound
		}
		contract = copyMap(ver.Contract)
		source = ContractSourceTemplateRef
		if len(req.InlineContract) > 0 {
			mergeInto(contract, req.InlineContract)
			source = ContractSourceTemplateRefWithOverride
		}
		snap.TemplateID = req.TemplateID
		snap.TemplateVersion = version
	}

	if len(contract) == 0 {
		contract = map[string]any{}
	}

	snap.ContractSource = source
	snap.ContractSnapshot = contract
	s.taskContracts[taskKey] = snap
	return snap, nil
}

func (s *Service) CreateCritic(_ context.Context, req CreateCriticRequest) (Critic, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.CriticID) == "" || strings.TrimSpace(req.Name) == "" {
		return Critic{}, fmt.Errorf("%w: tenant_id, critic_id, and name are required", ErrInvalidRequest)
	}
	if !isStageName(req.StageName) {
		return Critic{}, fmt.Errorf("%w: invalid stage_name", ErrInvalidRequest)
	}
	if strings.TrimSpace(req.TaskFamily) == "" || strings.TrimSpace(req.CriticClass) == "" {
		return Critic{}, fmt.Errorf("%w: task_family and critic_class are required", ErrInvalidRequest)
	}
	if strings.TrimSpace(req.ProviderFamily) == "" || strings.TrimSpace(req.FailureModeClass) == "" {
		return Critic{}, fmt.Errorf("%w: provider_family and failure_mode_class are required", ErrInvalidRequest)
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	now := time.Now().UTC()
	ck := tenantScoped(req.TenantID, req.CriticID)

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.critics[ck]; ok {
		return Critic{}, fmt.Errorf("%w: critic_id", ErrAlreadyExists)
	}

	critic := Critic{
		TenantID:         strings.TrimSpace(req.TenantID),
		CriticID:         strings.TrimSpace(req.CriticID),
		Name:             strings.TrimSpace(req.Name),
		StageName:        strings.TrimSpace(req.StageName),
		TaskFamily:       strings.TrimSpace(req.TaskFamily),
		CriticClass:      strings.TrimSpace(req.CriticClass),
		ProviderFamily:   strings.TrimSpace(req.ProviderFamily),
		FailureModeClass: strings.TrimSpace(req.FailureModeClass),
		Enabled:          enabled,
		Status:           StatusDraft,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.critics[ck] = critic
	return critic, nil
}

func (s *Service) ListCritics(_ context.Context, tenantID string) []Critic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Critic, 0)
	for _, critic := range s.critics {
		if critic.TenantID != tenantID {
			continue
		}
		items = append(items, critic)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CriticID < items[j].CriticID
	})
	return items
}

func (s *Service) GetCritic(_ context.Context, tenantID, criticID string) (Critic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	critic, ok := s.critics[tenantScoped(tenantID, criticID)]
	if !ok {
		return Critic{}, ErrNotFound
	}
	return critic, nil
}

func (s *Service) CreateCriticVersion(_ context.Context, tenantID, criticID string, req CreateCriticVersionRequest) (CriticVersion, error) {
	if req.Version <= 0 {
		return CriticVersion{}, fmt.Errorf("%w: version must be > 0", ErrInvalidRequest)
	}
	if req.Config == nil {
		req.Config = map[string]any{}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.critics[tenantScoped(tenantID, criticID)]; !ok {
		return CriticVersion{}, ErrNotFound
	}

	vk := criticVersionKey(tenantID, criticID, req.Version)
	if _, ok := s.criticVersions[vk]; ok {
		return CriticVersion{}, fmt.Errorf("%w: critic version", ErrAlreadyExists)
	}

	ver := CriticVersion{
		TenantID:  tenantID,
		CriticID:  criticID,
		Version:   req.Version,
		Config:    copyMap(req.Config),
		CreatedAt: time.Now().UTC(),
	}
	s.criticVersions[vk] = ver
	return ver, nil
}

func (s *Service) PublishCritic(_ context.Context, tenantID, criticID string, version int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ck := tenantScoped(tenantID, criticID)
	critic, ok := s.critics[ck]
	if !ok {
		return ErrNotFound
	}
	if _, ok := s.criticVersions[criticVersionKey(tenantID, criticID, version)]; !ok {
		return ErrNotFound
	}
	critic.Status = StatusPublished
	critic.PublishedVersion = version
	critic.UpdatedAt = time.Now().UTC()
	s.critics[ck] = critic
	return nil
}

func (s *Service) StartValidationRun(_ context.Context, req StartValidationRunRequest) (ValidationRun, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.TaskID) == "" {
		return ValidationRun{}, fmt.Errorf("%w: tenant_id and task_id are required", ErrInvalidRequest)
	}
	if strings.TrimSpace(req.TaskFamily) == "" {
		return ValidationRun{}, fmt.Errorf("%w: task_family is required", ErrInvalidRequest)
	}
	if req.ValidationTier == "" {
		req.ValidationTier = ValidationTierStandard
	}
	if !isValidationTier(req.ValidationTier) {
		return ValidationRun{}, fmt.Errorf("%w: invalid validation_tier", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	contract, ok := s.taskContracts[tenantScoped(req.TenantID, req.TaskID)]
	if !ok {
		var err error
		contract, err = s.attachTaskContractLocked(AttachTaskContractRequest{
			TenantID:       req.TenantID,
			TaskID:         req.TaskID,
			TemplateID:     req.TemplateID,
			TemplateVer:    req.TemplateVersion,
			ValidationTier: req.ValidationTier,
			InlineContract: req.InlineContract,
		})
		if err != nil {
			return ValidationRun{}, err
		}
	}

	if req.ValidationTier == ValidationTierHighAssurance {
		if err := s.ensureCriticDiversityLocked(req.TenantID, req.TaskFamily); err != nil {
			return ValidationRun{}, err
		}
	}

	stages := stagesForTier(req.ValidationTier)
	results := make([]ValidationStageResult, 0, len(stages))
	decision := DecisionAccepted
	status := RunStatusCompleted
	now := time.Now().UTC()
	seed := fmt.Sprintf("%s:%s:%d", req.TaskID, req.ResultReference, len(stages))

	for idx, stage := range stages {
		critic, hasCritic := s.pickCriticLocked(req.TenantID, req.TaskFamily, stage, idx)
		score := deterministicScore(seed + ":" + stage)
		stageDecision := DecisionAccepted
		if !hasCritic {
			score = 0.5
			stageDecision = DecisionNeedsReview
		} else {
			threshold := 0.64
			if req.ValidationTier == ValidationTierHighAssurance {
				threshold = 0.72
			}
			if score < threshold {
				stageDecision = DecisionNeedsReview
			}
		}

		if stageDecision == DecisionRejected {
			decision = DecisionRejected
			status = RunStatusRejected
		} else if stageDecision == DecisionNeedsReview && status != RunStatusRejected {
			decision = DecisionNeedsReview
			status = RunStatusNeedsReview
		}

		stageResult := ValidationStageResult{
			StageOrder:   idx + 1,
			StageName:    stage,
			Decision:     stageDecision,
			Score:        score,
			OccurredAt:   now,
			EvidenceRefs: []string{"evidence://" + req.TaskID + "/" + stage},
		}
		if hasCritic {
			stageResult.CriticID = critic.CriticID
			stageResult.CriticClass = critic.CriticClass
		}
		results = append(results, stageResult)
	}

	runID := s.nextIDLocked("val")
	run := ValidationRun{
		ValidationRunID: runID,
		TenantID:        req.TenantID,
		TaskID:          req.TaskID,
		TaskFamily:      req.TaskFamily,
		ResultReference: req.ResultReference,
		ValidationTier:  req.ValidationTier,
		Status:          status,
		Decision:        decision,
		Contract:        contract,
		StageResults:    results,
		CreatedAt:       now,
		CompletedAt:     now,
	}

	s.validationRuns[tenantScoped(req.TenantID, runID)] = run
	return run, nil
}

func (s *Service) ListTaskValidationRuns(_ context.Context, tenantID, taskID string) []ValidationRun {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]ValidationRun, 0)
	for _, run := range s.validationRuns {
		if run.TenantID == tenantID && run.TaskID == taskID {
			items = append(items, run)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *Service) ListValidationRuns(_ context.Context, tenantID string) []ValidationRun {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]ValidationRun, 0)
	for _, run := range s.validationRuns {
		if run.TenantID != tenantID {
			continue
		}
		items = append(items, run)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *Service) GetValidationRun(_ context.Context, tenantID, validationRunID string) (ValidationRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	run, ok := s.validationRuns[tenantScoped(tenantID, validationRunID)]
	if !ok {
		return ValidationRun{}, ErrNotFound
	}
	return run, nil
}

func (s *Service) ListValidationStages(_ context.Context, tenantID, validationRunID string) ([]ValidationStageResult, error) {
	run, err := s.GetValidationRun(context.Background(), tenantID, validationRunID)
	if err != nil {
		return nil, err
	}
	return run.StageResults, nil
}

func (s *Service) EscalateValidationRun(_ context.Context, req EscalateValidationRunRequest) (ValidationRun, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ValidationRunID) == "" || strings.TrimSpace(req.Reason) == "" {
		return ValidationRun{}, fmt.Errorf("%w: tenant_id, validation_run_id, and reason are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	runKey := tenantScoped(req.TenantID, req.ValidationRunID)
	run, ok := s.validationRuns[runKey]
	if !ok {
		return ValidationRun{}, ErrNotFound
	}
	run.Status = RunStatusEscalated
	run.Decision = DecisionNeedsReview
	run.CompletedAt = time.Now().UTC()
	s.validationRuns[runKey] = run

	esc := ValidationEscalation{
		TenantID:         req.TenantID,
		ValidationRunID:  req.ValidationRunID,
		Reason:           strings.TrimSpace(req.Reason),
		EscalatedByActor: strings.TrimSpace(req.EscalatedByActor),
		CreatedAt:        time.Now().UTC(),
	}
	s.escalations[runKey] = append(s.escalations[runKey], esc)
	return run, nil
}

func (s *Service) ValidationDashboard(_ context.Context, tenantID string, limit int) ValidationDashboard {
	if limit <= 0 {
		limit = 20
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	statusCounts := map[string]int{}
	decisionCounts := map[string]int{}
	tierCounts := map[string]int{}
	stageIssueCounts := map[string]int{}
	runs := make([]ValidationRun, 0)

	for _, run := range s.validationRuns {
		if run.TenantID != tenantID {
			continue
		}
		runs = append(runs, run)
		statusCounts[run.Status]++
		decisionCounts[run.Decision]++
		tierCounts[run.ValidationTier]++
		for _, stage := range run.StageResults {
			if stage.Decision != DecisionAccepted {
				stageIssueCounts[stage.StageName]++
			}
		}
	}
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].CreatedAt.After(runs[j].CreatedAt)
	})
	recentRuns := runs
	if len(recentRuns) > limit {
		recentRuns = recentRuns[:limit]
	}

	escalations := make([]ValidationEscalation, 0)
	for key, runEscalations := range s.escalations {
		if !strings.HasPrefix(key, strings.TrimSpace(tenantID)+"::") {
			continue
		}
		escalations = append(escalations, runEscalations...)
	}
	sort.Slice(escalations, func(i, j int) bool {
		return escalations[i].CreatedAt.After(escalations[j].CreatedAt)
	})
	queue := escalations
	if len(queue) > limit {
		queue = queue[:limit]
	}

	residualRate := 0.0
	if len(runs) > 0 {
		residualRate = float64(len(escalations)) / float64(len(runs))
	}

	return ValidationDashboard{
		TenantID:               tenantID,
		TotalRuns:              len(runs),
		StatusCounts:           statusCounts,
		DecisionCounts:         decisionCounts,
		TierCounts:             tierCounts,
		StageIssueCounts:       stageIssueCounts,
		RecentRuns:             recentRuns,
		EscalationQueue:        queue,
		EscalationResidualRate: residualRate,
	}
}

func (s *Service) attachTaskContractLocked(req AttachTaskContractRequest) (TaskAcceptanceContract, error) {
	if req.ValidationTier == "" {
		req.ValidationTier = ValidationTierStandard
	}
	if !isValidationTier(req.ValidationTier) {
		return TaskAcceptanceContract{}, fmt.Errorf("%w: invalid validation_tier", ErrInvalidRequest)
	}

	taskKey := tenantScoped(req.TenantID, req.TaskID)
	if existing, ok := s.taskContracts[taskKey]; ok {
		return existing, nil
	}

	snap := TaskAcceptanceContract{
		TenantID:       req.TenantID,
		TaskID:         req.TaskID,
		ValidationTier: req.ValidationTier,
		CreatedAt:      time.Now().UTC(),
	}

	contract := copyMap(req.InlineContract)
	source := ContractSourceInline
	if req.TemplateID != "" {
		tpl, ok := s.templates[tenantScoped(req.TenantID, req.TemplateID)]
		if !ok {
			return TaskAcceptanceContract{}, ErrNotFound
		}
		version := req.TemplateVer
		if version <= 0 {
			version = tpl.PublishedVersion
		}
		if version <= 0 {
			return TaskAcceptanceContract{}, ErrTemplateNotReady
		}
		ver, ok := s.templateVersions[templateVersionKey(req.TenantID, req.TemplateID, version)]
		if !ok {
			return TaskAcceptanceContract{}, ErrNotFound
		}
		contract = copyMap(ver.Contract)
		source = ContractSourceTemplateRef
		if len(req.InlineContract) > 0 {
			mergeInto(contract, req.InlineContract)
			source = ContractSourceTemplateRefWithOverride
		}
		snap.TemplateID = req.TemplateID
		snap.TemplateVersion = version
	}

	if len(contract) == 0 {
		contract = map[string]any{}
	}
	snap.ContractSource = source
	snap.ContractSnapshot = contract
	s.taskContracts[taskKey] = snap
	return snap, nil
}

func (s *Service) ensureCriticDiversityLocked(tenantID, taskFamily string) error {
	providerFamilies := map[string]struct{}{}
	failureClasses := map[string]struct{}{}
	count := 0
	for _, critic := range s.critics {
		if critic.TenantID != tenantID || !critic.Enabled || critic.Status != StatusPublished {
			continue
		}
		if critic.TaskFamily != taskFamily && critic.TaskFamily != "*" {
			continue
		}
		count++
		providerFamilies[critic.ProviderFamily] = struct{}{}
		failureClasses[critic.FailureModeClass] = struct{}{}
	}
	if count < 2 || len(providerFamilies) < 2 || len(failureClasses) < 2 {
		return ErrCriticDiversity
	}
	return nil
}

func (s *Service) pickCriticLocked(tenantID, taskFamily, stage string, offset int) (Critic, bool) {
	candidates := make([]Critic, 0)
	for _, critic := range s.critics {
		if critic.TenantID != tenantID || !critic.Enabled || critic.Status != StatusPublished {
			continue
		}
		if critic.TaskFamily != taskFamily && critic.TaskFamily != "*" {
			continue
		}
		if critic.StageName != stage {
			continue
		}
		candidates = append(candidates, critic)
	}
	if len(candidates) == 0 {
		return Critic{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].CriticID < candidates[j].CriticID
	})
	return candidates[offset%len(candidates)], true
}

func (s *Service) nextIDLocked(prefix string) string {
	s.seq++
	return fmt.Sprintf("%s_%06d", prefix, s.seq)
}

func isValidationTier(v string) bool {
	switch strings.TrimSpace(v) {
	case ValidationTierFast, ValidationTierStandard, ValidationTierHighAssurance:
		return true
	default:
		return false
	}
}

func isStageName(v string) bool {
	switch strings.TrimSpace(v) {
	case StagePlan, StageExecution, StageArtifact, StageOutput, StagePolicy, StageSecurity, StageBenchmark, StageReduction:
		return true
	default:
		return false
	}
}

func stagesForTier(v string) []string {
	switch strings.TrimSpace(v) {
	case ValidationTierFast:
		return []string{StagePlan, StageOutput}
	case ValidationTierHighAssurance:
		return []string{StagePlan, StageExecution, StageArtifact, StageOutput, StagePolicy, StageSecurity, StageBenchmark, StageReduction}
	default:
		return []string{StagePlan, StageExecution, StageOutput, StagePolicy, StageSecurity}
	}
}

func deterministicScore(seed string) float64 {
	sum := sha1.Sum([]byte(seed))
	raw := float64(int(sum[0])+int(sum[1])+int(sum[2])) / (255.0 * 3.0)
	return 0.55 + (raw * 0.4)
}

func tenantScoped(tenantID, id string) string {
	return strings.TrimSpace(tenantID) + "::" + strings.TrimSpace(id)
}

func templateVersionKey(tenantID, templateID string, version int) string {
	return fmt.Sprintf("%s::%s::%d", strings.TrimSpace(tenantID), strings.TrimSpace(templateID), version)
}

func criticVersionKey(tenantID, criticID string, version int) string {
	return fmt.Sprintf("%s::%s::%d", strings.TrimSpace(tenantID), strings.TrimSpace(criticID), version)
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

func mergeInto(dst map[string]any, src map[string]any) {
	for k, v := range src {
		dst[k] = v
	}
}

func HashContract(contract map[string]any) string {
	b, _ := json.Marshal(contract)
	sum := sha1.Sum(b)
	return hex.EncodeToString(sum[:])
}
