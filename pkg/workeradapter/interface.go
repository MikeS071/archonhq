package workeradapter

import "context"

type NodeCapabilities struct{}
type Lease struct{}
type TaskSpec struct{}
type RunHandle struct{}
type RunStatus struct{}
type ResultBundle struct{}

type Adapter interface {
	Name() string
	Capabilities(ctx context.Context) (NodeCapabilities, error)
	StartLease(ctx context.Context, lease Lease, task TaskSpec) (RunHandle, error)
	PollRun(ctx context.Context, handle RunHandle) (RunStatus, error)
	CollectResult(ctx context.Context, handle RunHandle) (ResultBundle, error)
	CancelRun(ctx context.Context, handle RunHandle) error
}
