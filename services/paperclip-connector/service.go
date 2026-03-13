package paperclipconnector

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MikeS071/archonhq/integrations/paperclip"
)

type SyncRequest struct {
	SyncID      string
	InitiatedBy string
	DryRun      bool
	Force       bool
	Payload     paperclip.ProjectionPayload
}

type SyncResult struct {
	SyncID        string
	Status        string
	ExternalRef   string
	GeneratedAt   time.Time
	SourceOfTruth string
	SurfaceCounts map[string]int
}

type Service struct {
	connector paperclip.Connector
	now       func() time.Time
}

func New(connector paperclip.Connector) *Service {
	if connector == nil {
		connector = paperclip.NoopConnector{}
	}
	return &Service{
		connector: connector,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Sync(ctx context.Context, req SyncRequest) (SyncResult, error) {
	syncID := strings.TrimSpace(req.SyncID)
	if syncID == "" {
		return SyncResult{}, errors.New("sync_id is required")
	}
	if strings.TrimSpace(req.InitiatedBy) == "" {
		return SyncResult{}, errors.New("initiated_by is required")
	}
	if strings.TrimSpace(req.Payload.TenantID) == "" {
		return SyncResult{}, errors.New("tenant_id is required")
	}
	if req.Payload.SourceOfTruth != "postgres" {
		return SyncResult{}, errors.New("paperclip sync source_of_truth must be postgres")
	}

	generatedAt := req.Payload.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = s.now()
		req.Payload.GeneratedAt = generatedAt
	}

	result := SyncResult{
		SyncID:        syncID,
		Status:        "completed",
		GeneratedAt:   generatedAt,
		SourceOfTruth: req.Payload.SourceOfTruth,
		SurfaceCounts: map[string]int{
			"workspace_summaries": len(req.Payload.WorkspaceSummaries),
			"approvals":           len(req.Payload.Approvals),
			"fleet":               len(req.Payload.Fleet),
			"reliability":         len(req.Payload.Reliability),
			"settlements":         len(req.Payload.Settlements),
		},
	}

	if req.DryRun {
		result.Status = "dry_run"
		result.ExternalRef = "paperclip://dry-run/" + syncID
		return result, nil
	}

	externalRef, err := s.connector.PushProjection(ctx, req.Payload)
	if err != nil {
		return SyncResult{}, err
	}
	result.ExternalRef = externalRef
	return result, nil
}
