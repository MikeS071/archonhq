package hermesadapter

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/MikeS071/archonhq/pkg/policy"
	"github.com/MikeS071/archonhq/pkg/workeradapter"
)

// Adapter is the production runtime adapter in v1.
type Adapter struct {
	mu   sync.Mutex
	runs map[string]runRecord
}

type runRecord struct {
	lease     workeradapter.Lease
	task      workeradapter.TaskSpec
	startedAt time.Time
	state     string
	backend   string
}

func New() *Adapter {
	return &Adapter{
		runs: map[string]runRecord{},
	}
}

func (a *Adapter) Name() string {
	return "hermes"
}

func (a *Adapter) Capabilities(context.Context) (workeradapter.NodeCapabilities, error) {
	return workeradapter.NodeCapabilities{
		RuntimeName:      "hermes",
		AllowedBackends:  []string{"docker", "ssh", "modal"},
		AllowedToolsets:  []string{"file", "terminal", "web", "mcp"},
		RequiresBYOKOnly: true,
	}, nil
}

func (a *Adapter) StartLease(_ context.Context, lease workeradapter.Lease, task workeradapter.TaskSpec) (workeradapter.RunHandle, error) {
	if lease.LeaseID == "" {
		return workeradapter.RunHandle{}, fmt.Errorf("lease_id is required")
	}
	if lease.ByokProviderRef == "" {
		return workeradapter.RunHandle{}, fmt.Errorf("byok_provider_ref is required")
	}
	execPolicy, err := policy.NormalizeExecutionPolicy(lease.ExecutionPolicy)
	if err != nil {
		return workeradapter.RunHandle{}, fmt.Errorf("invalid execution policy: %w", err)
	}

	runID := fmt.Sprintf("run_%d", time.Now().UTC().UnixNano())
	workspacePath := filepath.Join("/tmp", "archonhq", "runs", lease.TenantID, lease.LeaseID, runID)
	handle := workeradapter.RunHandle{
		RunID:          runID,
		LeaseID:        lease.LeaseID,
		WorkspacePath:  workspacePath,
		PrimaryBackend: execPolicy.AllowedBackends[0],
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.runs[runID] = runRecord{
		lease:     lease,
		task:      task,
		startedAt: time.Now().UTC(),
		state:     "running",
		backend:   execPolicy.AllowedBackends[0],
	}
	return handle, nil
}

func (a *Adapter) PollRun(_ context.Context, handle workeradapter.RunHandle) (workeradapter.RunStatus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	run, ok := a.runs[handle.RunID]
	if !ok {
		return workeradapter.RunStatus{}, fmt.Errorf("run not found")
	}
	status := workeradapter.RunStatus{
		RunID:     handle.RunID,
		State:     run.state,
		StartedAt: run.startedAt,
	}
	if run.state == "completed" || run.state == "cancelled" {
		finishedAt := time.Now().UTC()
		status.FinishedAt = &finishedAt
	}
	return status, nil
}

func (a *Adapter) CollectResult(_ context.Context, handle workeradapter.RunHandle) (workeradapter.ResultBundle, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	run, ok := a.runs[handle.RunID]
	if !ok {
		return workeradapter.ResultBundle{}, fmt.Errorf("run not found")
	}
	run.state = "completed"
	a.runs[handle.RunID] = run

	return workeradapter.ResultBundle{
		ResultID:            "res_" + handle.RunID,
		OutputRefs:          []string{fmt.Sprintf("art_%s_out", handle.RunID)},
		LogsArtifactID:      fmt.Sprintf("art_%s_logs", handle.RunID),
		ToolCallsArtifactID: fmt.Sprintf("art_%s_tools", handle.RunID),
		MetricsArtifactID:   fmt.Sprintf("art_%s_metrics", handle.RunID),
		Signature:           fmt.Sprintf("signed:%s:%s:%s", run.lease.NodeID, run.lease.LeaseID, "res_"+handle.RunID),
	}, nil
}

func (a *Adapter) CancelRun(_ context.Context, handle workeradapter.RunHandle) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	run, ok := a.runs[handle.RunID]
	if !ok {
		return fmt.Errorf("run not found")
	}
	run.state = "cancelled"
	a.runs[handle.RunID] = run
	return nil
}
