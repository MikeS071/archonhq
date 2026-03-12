package objectstore

import "testing"

func TestNewAndHealth(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatalf("expected validation error")
	}

	c, err := New(Config{Endpoint: "http://minio:9000", Bucket: "archonhq"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := c.Health(); err != nil {
		t.Fatalf("health: %v", err)
	}
}
