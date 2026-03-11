# HERMES_ADAPTER_SPEC.md

## Purpose
Bridge hub leases/tasks into Hermes runtime execution.

## Runtime assumptions from public Hermes positioning
- one-line install / easy node join
- tool support
- Docker / SSH / Modal backends
- command approval/safety model
- MCP support
- long-lived memory model

## Adapter requirements
- map TaskSpec -> Hermes prompt/context/tool grants/backend config
- create isolated per-task workspace
- prevent write-back to long-lived Hermes memory by default
- collect logs, tool calls, file changes, metrics
- upload artifacts
- submit signed results

## Go interface
```go
type WorkerAdapter interface {
    Name() string
    Capabilities(ctx context.Context) (NodeCapabilities, error)
    StartLease(ctx context.Context, lease Lease, task TaskSpec) (RunHandle, error)
    PollRun(ctx context.Context, handle RunHandle) (RunStatus, error)
    CollectResult(ctx context.Context, handle RunHandle) (ResultBundle, error)
    CancelRun(ctx context.Context, handle RunHandle) error
}
```
