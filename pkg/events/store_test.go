package events

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestAppendSuccessAndMarshalFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewPostgresStore(db)
	if store == nil {
		t.Fatalf("expected store")
	}

	env := Envelope{
		EventID:      "evt_1",
		TenantID:     "ten_1",
		EntityType:   "task",
		EntityID:     "task_1",
		EventType:    "task.created",
		EventVersion: 1,
		Payload:      map[string]any{"k": "v"},
		OccurredAt:   time.Now().UTC(),
	}

	mock.ExpectExec("INSERT INTO event_records").
		WithArgs(env.EventID, env.TenantID, env.WorkspaceID, env.EntityType, env.EntityID, env.EventType, env.EventVersion, env.ActorType, env.ActorID, env.CorrelationID, env.IdempotencyKey, sqlmock.AnyArg(), env.OccurredAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := store.Append(context.Background(), env); err != nil {
		t.Fatalf("append: %v", err)
	}

	badEnv := env
	badEnv.Payload = map[string]any{"bad": make(chan int)}
	if err := store.Append(context.Background(), badEnv); err == nil {
		t.Fatalf("expected marshal failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
