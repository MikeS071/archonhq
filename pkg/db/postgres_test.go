package db

import (
	"context"
	"testing"
	"time"
)

func TestOpenFailurePath(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := Open(ctx, "postgres://invalid:invalid@127.0.0.1:1/archonhq?sslmode=disable&connect_timeout=1"); err == nil {
		t.Fatalf("expected open/ping failure")
	}
}
