package nats

import (
	"context"
	"testing"
)

func TestHealthAndConnectFailures(t *testing.T) {
	c := &Client{}
	if err := c.Health(context.Background()); err == nil {
		t.Fatalf("expected health error for disconnected client")
	}
	c.Close() // no panic

	if _, err := Connect("::://bad-url"); err == nil {
		t.Fatalf("expected connect error")
	}
}
