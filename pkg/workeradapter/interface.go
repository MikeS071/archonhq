package workeradapter

import (
	"context"
	"time"
)

type NodeCapabilities struct {
	RuntimeName      string
	AllowedBackends  []string
	AllowedToolsets  []string
	RequiresBYOKOnly bool
}

type Lease struct {
	LeaseID          string
	TenantID         string
	TaskID           string
	NodeID           string
	WorkspaceRootRef string
	ByokProviderRef  string
	ExecutionPolicy  map[string]any
}

type TaskSpec struct {
	TaskID       string
	WorkspaceID  string
	TaskFamily   string
	Title        string
	Description  string
	InputRefs    []string
	SchemaRef    string
	ApprovalMode string
}

type RunHandle struct {
	RunID          string
	LeaseID        string
	WorkspacePath  string
	PrimaryBackend string
}

type RunStatus struct {
	RunID      string
	State      string
	StartedAt  time.Time
	FinishedAt *time.Time
}

type ResultBundle struct {
	ResultID            string
	OutputRefs          []string
	LogsArtifactID      string
	ToolCallsArtifactID string
	MetricsArtifactID   string
	Signature           string
}

type Adapter interface {
	Name() string
	Capabilities(ctx context.Context) (NodeCapabilities, error)
	StartLease(ctx context.Context, lease Lease, task TaskSpec) (RunHandle, error)
	PollRun(ctx context.Context, handle RunHandle) (RunStatus, error)
	CollectResult(ctx context.Context, handle RunHandle) (ResultBundle, error)
	CancelRun(ctx context.Context, handle RunHandle) error
}
