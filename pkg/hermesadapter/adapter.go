package hermesadapter

import (
	"context"
	"fmt"

	"github.com/MikeS071/archonhq/pkg/workeradapter"
)

// Adapter is the production runtime adapter in v1.
type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Name() string {
	return "hermes"
}

func (a *Adapter) Capabilities(context.Context) (workeradapter.NodeCapabilities, error) {
	return workeradapter.NodeCapabilities{}, nil
}

func (a *Adapter) StartLease(context.Context, workeradapter.Lease, workeradapter.TaskSpec) (workeradapter.RunHandle, error) {
	return workeradapter.RunHandle{}, fmt.Errorf("not implemented")
}

func (a *Adapter) PollRun(context.Context, workeradapter.RunHandle) (workeradapter.RunStatus, error) {
	return workeradapter.RunStatus{}, fmt.Errorf("not implemented")
}

func (a *Adapter) CollectResult(context.Context, workeradapter.RunHandle) (workeradapter.ResultBundle, error) {
	return workeradapter.ResultBundle{}, fmt.Errorf("not implemented")
}

func (a *Adapter) CancelRun(context.Context, workeradapter.RunHandle) error {
	return fmt.Errorf("not implemented")
}
