package nats

import (
	"context"
	"testing"

	natsgo "github.com/nats-io/nats.go"
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

func TestHealthWithDisconnectedNonNilConn(t *testing.T) {
	c := &Client{Conn: &natsgo.Conn{}}
	if err := c.Health(context.Background()); err == nil {
		t.Fatalf("expected health error for disconnected conn struct")
	}
}
