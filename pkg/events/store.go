package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Store interface {
	Append(ctx context.Context, env Envelope) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Append(ctx context.Context, env Envelope) error {
	payload, err := json.Marshal(env.Payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	const q = `
INSERT INTO event_records (
  event_id,
  tenant_id,
  workspace_id,
  entity_type,
  entity_id,
  event_type,
  event_version,
  actor_type,
  actor_id,
  correlation_id,
  idempotency_key,
  payload_json,
  occurred_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

	_, err = s.db.ExecContext(
		ctx,
		q,
		env.EventID,
		env.TenantID,
		env.WorkspaceID,
		env.EntityType,
		env.EntityID,
		env.EventType,
		env.EventVersion,
		env.ActorType,
		env.ActorID,
		env.CorrelationID,
		env.IdempotencyKey,
		payload,
		env.OccurredAt,
	)
	if err != nil {
		return fmt.Errorf("insert event record: %w", err)
	}

	return nil
}
