package events

import "time"

// Envelope matches docs/EVENT_CATALOG.md.
type Envelope struct {
	EventID        string         `json:"event_id"`
	TenantID       string         `json:"tenant_id"`
	WorkspaceID    string         `json:"workspace_id,omitempty"`
	EntityType     string         `json:"entity_type"`
	EntityID       string         `json:"entity_id"`
	EventType      string         `json:"event_type"`
	EventVersion   int            `json:"event_version"`
	ActorType      string         `json:"actor_type,omitempty"`
	ActorID        string         `json:"actor_id,omitempty"`
	CorrelationID  string         `json:"correlation_id,omitempty"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
	Payload        map[string]any `json:"payload"`
	OccurredAt     time.Time      `json:"occurred_at"`
}
