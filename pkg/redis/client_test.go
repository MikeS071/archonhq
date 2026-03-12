package redis

import "testing"

func TestConnectAndHealth(t *testing.T) {
	if _, err := Connect(""); err == nil {
		t.Fatalf("expected connect validation error")
	}
	c, err := Connect("redis://localhost:6379")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := c.Health(); err != nil {
		t.Fatalf("health: %v", err)
	}
}
