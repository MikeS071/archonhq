package paperclip

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ProjectionPayload struct {
	TenantID           string           `json:"tenant_id"`
	GeneratedAt        time.Time        `json:"generated_at"`
	SourceOfTruth      string           `json:"source_of_truth"`
	WorkspaceSummaries []map[string]any `json:"workspace_summaries"`
	Approvals          []map[string]any `json:"approvals"`
	Fleet              []map[string]any `json:"fleet"`
	Reliability        []map[string]any `json:"reliability"`
	Settlements        []map[string]any `json:"settlements"`
	Metadata           map[string]any   `json:"metadata,omitempty"`
}

type Connector interface {
	PushProjection(ctx context.Context, payload ProjectionPayload) (externalRef string, err error)
}

type NoopConnector struct{}

func (NoopConnector) PushProjection(_ context.Context, payload ProjectionPayload) (string, error) {
	if strings.TrimSpace(payload.TenantID) == "" {
		return "", errors.New("tenant_id is required")
	}
	if payload.SourceOfTruth != "postgres" {
		return "", errors.New("paperclip projection source_of_truth must be postgres")
	}
	return fmt.Sprintf("paperclip_stub_%s", payload.GeneratedAt.UTC().Format("20060102T150405Z")), nil
}
