package db

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestHealthAndClose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}

	p := &Postgres{DB: db}
	if err := p.Health(context.Background()); err != nil {
		t.Fatalf("health: %v", err)
	}
	mock.ExpectClose()
	if err := p.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
